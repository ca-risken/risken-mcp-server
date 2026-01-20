package streamablehttp

import (
	"fmt"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
)

// ServeHTTP handles MCP requests(/mcp) with unified token validation.
// Automatically detects Project or Organization token.
func (a *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract requestID from JSON-RPC
	requestID, err := riskenmcp.ParseJSONRPCRequestID(r)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(nil, riskenmcp.JSONRPCErrorParseError, "Parse error(requestID)")
		http.Error(w, jsonRPCError.String(), http.StatusBadRequest)
		return
	}

	// Extract token from authorization header
	riskenToken := helper.ExtractRISKENTokenFromHeader(r)
	if riskenToken == "" {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Unauthorized(no authorization header)")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Auto-detect token type and create client
	unifiedClient, err := helper.DetectAndCreateClient(r.Context(), a.riskenURL, riskenToken)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, fmt.Sprintf("Invalid token: %s", err))
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add the appropriate client to the request context based on token type
	ctx := r.Context()
	switch unifiedClient.TokenType {
	case helper.TokenTypeOrganization:
		ctx = riskenmcp.WithOrganizationClient(ctx, unifiedClient.OrgClient)
	case helper.TokenTypeProject:
		ctx = riskenmcp.WithRISKENClient(ctx, unifiedClient.ProjectClient)
	}
	r = r.WithContext(ctx)

	// Delegate to the original handler
	a.StreamableHTTPServer.ServeHTTP(w, r)
}
