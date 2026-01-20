package streamablehttp

import (
	"net/http"

	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
)

// ServeHTTP handles MCP requests(/mcp) with unified token validation.
// Works with both Project and Organization tokens using a single client.
func (a *AuthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract requestID from JSON-RPC
	requestID, err := riskenmcp.ParseJSONRPCRequestID(r)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(nil, riskenmcp.JSONRPCErrorParseError, "Parse error(requestID)")
		http.Error(w, jsonRPCError.String(), http.StatusBadRequest)
		return
	}

	// Extract token from authorization header
	token := helper.ExtractRISKENTokenFromHeader(r)
	if token == "" {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Unauthorized(no authorization header)")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Create single RISKEN client (works with both Project and Organization tokens)
	riskenClient := risken.NewClient(token, risken.WithAPIEndpoint(a.riskenURL))

	// Signin to validate token
	if _, err := riskenClient.Signin(r.Context()); err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Unauthorized: invalid token")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add RISKEN Client to the request context
	ctx := riskenmcp.WithRISKENClient(r.Context(), riskenClient)
	r = r.WithContext(ctx)

	// Delegate to the original handler
	a.StreamableHTTPServer.ServeHTTP(w, r)
}
