package oauth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/mark3labs/mcp-go/server"
)

// Config for JWT validation against IdP
type Config struct {
	MCPServerURL string `json:"mcp_server_url"`

	// IdP
	AuthzMetadataEndpoint string `json:"authz_metadata_endpoint"`
	ClientID              string `json:"client_id"`
	ClientSecret          string `json:"client_secret"`
}

// Server implements MCP Resource Server with JWT validation
type Server struct {
	*server.StreamableHTTPServer
	config          *Config
	jwtValidator    *JWTValidator
	riskenURL       string
	mcpEndpointPath string
	logger          *slog.Logger
	httpServer      *http.Server

	// Cached OAuth2.1 metadata from IdP (read-only after initialization)
	oauth21Metadata *OAuth21Metadata
}

// NewServer creates MCP Resource Server with JWT validation
func NewServer(
	mcpServer *server.MCPServer,
	oauthConfig *Config,
	riskenURL, mcpEndpointPath string,
	logger *slog.Logger,
) *Server {
	jwtValidator := NewJWTValidator(oauthConfig.MCPServerURL, logger)

	return &Server{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath(mcpEndpointPath)),
		config:               oauthConfig,
		jwtValidator:         jwtValidator,
		riskenURL:            riskenURL,
		mcpEndpointPath:      mcpEndpointPath,
		logger:               logger,
	}
}

// Initialize loads JWKS and Authorization Server metadata from IdP
func (s *Server) Initialize(ctx context.Context) error {
	if s.config.ClientID == "" || s.config.ClientSecret == "" {
		return fmt.Errorf("client_id and client_secret are required")
	}

	// Load and validate authorization server metadata
	if err := s.LoadMetadata(ctx); err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Load JWKS for JWT validation
	if err := s.jwtValidator.LoadJWKS(ctx, s.oauth21Metadata); err != nil {
		return fmt.Errorf("failed to load JWKS: %w", err)
	}

	s.logger.Info("MCP Resource Server initialized successfully",
		slog.String("idp_issuer", s.oauth21Metadata.Issuer))

	return nil
}

// Start starts the integrated server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// MCP endpoint (OAuth protected)
	mux.Handle(s.mcpEndpointPath, s)

	// Metadata endpoints (Protected Resource only)
	mux.HandleFunc("/.well-known/oauth-protected-resource", s.handleProtectedResourceMetadata) // REQUIRED by MCP spec
	// Authorization Server Metadata proxy endpoint
	mux.HandleFunc("/.well-known/oauth-authorization-server", s.handleAuthorizationServerMetadata)

	// Dynamic Client Registration endpoint
	mux.HandleFunc("/register", s.handleDynamicClientRegistration)

	// Health check
	mux.HandleFunc("/health", s.healthzHandler)

	handler := helper.UseAccessLogging(s.logger)(mux)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// ServeHTTP handles MCP requests(/mcp) with OAuth token validation
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract requestID from JSON-RPC
	requestID, err := riskenmcp.ParseJSONRPCRequestID(r)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(nil, riskenmcp.JSONRPCErrorParseError, "Parse error(requestID)")
		http.Error(w, jsonRPCError.String(), http.StatusBadRequest)
		return
	}

	// Extract Bearer token
	token := helper.ExtractBearerToken(r)
	if token == "" {
		// Add WWW-Authenticate header as required by MCP spec (RFC9728 Section 5.1)
		metadataURL := s.config.MCPServerURL + "/.well-known/oauth-protected-resource"
		w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+metadataURL+`"`)
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Bearer token required")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Validate JWT token from IdP
	claims, err := s.jwtValidator.ValidateToken(token)
	if err != nil {
		// Add WWW-Authenticate header as required by MCP spec (RFC9728 Section 5.1)
		metadataURL := s.config.MCPServerURL + "/.well-known/oauth-protected-resource"
		w.Header().Set("WWW-Authenticate", `Bearer resource_metadata="`+metadataURL+`"`)
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Invalid JWT token")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Create RISKEN client for the user
	riskenToken := helper.ExtractRISKENTokenFromHeader(r)
	riskenClient, err := helper.CreateAndValidateRISKENClient(r.Context(), s.riskenURL, riskenToken)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorInternalError, "Failed to create RISKEN client")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add to context
	ctx := riskenmcp.WithRISKENClient(r.Context(), riskenClient)
	r = r.WithContext(ctx)

	// Log authenticated request
	s.logger.Debug("JWT authenticated request",
		slog.String("user", claims.Email),
		slog.String("username", claims.Username),
		slog.String("scope", claims.Scope))

	// Delegate to MCP server
	s.StreamableHTTPServer.ServeHTTP(w, r)
}
