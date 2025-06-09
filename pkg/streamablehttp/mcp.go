package streamablehttp

import (
	"fmt"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
)

// ServeHTTP handles MCP requests(/mcp) with RISKEN token validation
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

	// Verify token
	riskenClient, err := helper.CreateAndValidateRISKENClient(r.Context(), a.riskenURL, riskenToken)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, fmt.Sprintf("Invalid RISKEN token: %s", err))
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add RISKEN Client to the request context
	ctx := riskenmcp.WithRISKENClient(r.Context(), riskenClient)
	r = r.WithContext(ctx)

	// Delegate to the original handler
	a.StreamableHTTPServer.ServeHTTP(w, r)
}
