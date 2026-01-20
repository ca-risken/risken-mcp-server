package streamablehttp

import (
	"fmt"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
)

// ServeHTTP handles MCP requests(/mcp) with Organization token validation
func (a *OrgAuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract requestID from JSON-RPC
	requestID, err := riskenmcp.ParseJSONRPCRequestID(r)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(nil, riskenmcp.JSONRPCErrorParseError, "Parse error(requestID)")
		http.Error(w, jsonRPCError.String(), http.StatusBadRequest)
		return
	}

	// Extract token from authorization header
	orgToken := helper.ExtractRISKENTokenFromHeader(r)
	if orgToken == "" {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Unauthorized(no authorization header)")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Verify Organization token
	orgClient, err := helper.CreateAndValidateOrganizationClient(r.Context(), a.riskenURL, orgToken)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, fmt.Sprintf("Invalid Organization token: %s", err))
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add Organization Client to the request context
	ctx := riskenmcp.WithOrganizationClient(r.Context(), orgClient)
	r = r.WithContext(ctx)

	// Delegate to the original handler
	a.StreamableHTTPServer.ServeHTTP(w, r)
}
