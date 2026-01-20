package oauth

import (
	"log/slog"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
)

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
