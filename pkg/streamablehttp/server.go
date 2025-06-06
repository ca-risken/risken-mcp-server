package streamablehttp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/mark3labs/mcp-go/server"
)

// AuthServer is a wrapper for StreamableHTTPServer with authentication
type AuthServer struct {
	*server.StreamableHTTPServer
	riskenURL    string
	endpointPath string
	logger       *slog.Logger
	httpServer   *http.Server
	mu           sync.RWMutex
}

// NewAuthServer creates a new authenticated server instance
func NewAuthServer(mcpServer *server.MCPServer, riskenURL, endpointPath string, logger *slog.Logger) *AuthServer {
	return &AuthServer{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath(endpointPath)),
		endpointPath:         endpointPath,
		riskenURL:            riskenURL,
		logger:               logger,
	}
}

// Override Start method to apply authentication
func (a *AuthServer) Start(addr string) error {
	a.mu.Lock()
	mux := http.NewServeMux()
	mux.Handle(a.endpointPath, a)
	mux.HandleFunc("/health", a.healthzHandler)
	handler := helper.UseAccessLogging(a.logger)(mux)

	a.httpServer = &http.Server{
		Addr:        addr,
		Handler:     handler,
		ReadTimeout: 300 * time.Second,
	}
	a.mu.Unlock()
	return a.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (a *AuthServer) Shutdown(ctx context.Context) error {
	a.mu.RLock()
	srv := a.httpServer
	a.mu.RUnlock()
	if srv != nil {
		return srv.Shutdown(ctx)
	}
	return nil
}

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
	token := helper.ExtractBearerToken(r)
	if token == "" {
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, "Unauthorized(no authorization header)")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Verify token
	riskenClient, err := helper.CreateAndValidateRISKENClient(r.Context(), a.riskenURL, token)
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
