package oauth

import (
	"log/slog"
	"net/http"

	"github.com/ca-risken/go-risken"
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

	// Verify token
	riskenToken := helper.ExtractRISKENTokenFromHeader(r)
	riskenClient := risken.NewClient(riskenToken, risken.WithAPIEndpoint(s.riskenURL))

	// Signin to validate token
	signinResp, err := riskenClient.Signin(r.Context())
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Failed to validate RISKEN token")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Log with token type info
	if signinResp.OrganizationID > 0 {
		s.logger.Debug("JWT authenticated request (Organization token)",
			slog.String("user", claims.Email),
			slog.String("username", claims.Username),
			slog.String("scope", claims.Scope),
			slog.Uint64("organization_id", uint64(signinResp.OrganizationID)))
	} else {
		s.logger.Debug("JWT authenticated request (Project token)",
			slog.String("user", claims.Email),
			slog.String("username", claims.Username),
			slog.String("scope", claims.Scope),
			slog.Uint64("project_id", uint64(signinResp.ProjectID)))
	}

	// Add RISKEN Client to context
	ctx := riskenmcp.WithRISKENClient(r.Context(), riskenClient)
	r = r.WithContext(ctx)

	// Delegate to MCP server
	s.StreamableHTTPServer.ServeHTTP(w, r)
}
