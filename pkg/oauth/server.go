package oauth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/server"
)

// Config for JWT validation against IdP
type Config struct {
	MCPServerURL string `json:"mcp_server_url" validate:"required,url"`

	// IdP Metadata Endpoint
	AuthzMetadataEndpoint string `json:"authz_metadata_endpoint" validate:"required,url"`

	// M2M Application credentials for DCR
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`

	// JWT Session signing key
	JWTSigningKey string `json:"jwt_signing_key" validate:"required"`
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

	// Cached OAuth2.1 metadata from IdP
	oauth21Metadata *OAuth21Metadata
	// Session manager for Third-Party Authorization Flow
	sessionManager SessionManager
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

// Start starts the integrated server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// MCP endpoint (default: /mcp)
	mux.Handle(s.mcpEndpointPath, s)

	// Third-Party Authorization Flow endpoints
	mux.HandleFunc("/authorize", s.handleAuthorize)          // Authorization endpoint
	mux.HandleFunc("/token", s.handleToken)                  // Token endpoint
	mux.HandleFunc("/register", s.handleRegister)            // Dynamic Client Registration
	mux.HandleFunc("/oauth/callback", s.handleOAuthCallback) // OAuth callback from IdP

	// Metadata endpoints
	mux.HandleFunc("/.well-known/oauth-protected-resource", s.handleProtectedResourceMetadata) // REQUIRED by MCP spec
	mux.HandleFunc("/.well-known/oauth-authorization-server", s.handleAuthorizationServerMetadata)

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
