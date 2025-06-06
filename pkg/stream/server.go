package stream

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

// AuthStreamableHTTPServer is a wrapper for StreamableHTTPServer with authentication
type AuthStreamableHTTPServer struct {
	*server.StreamableHTTPServer
	riskenURL    string
	endpointPath string
	logger       *slog.Logger
	httpServer   *http.Server
	mu           sync.RWMutex
}

// NewAutStreamableHTTPServer creates a new authenticated server instance
func NewAuthStreamableHTTPServer(mcpServer *server.MCPServer, riskenURL, endpointPath string, logger *slog.Logger) *AuthStreamableHTTPServer {
	return &AuthStreamableHTTPServer{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath(endpointPath)),
		endpointPath:         endpointPath,
		riskenURL:            riskenURL,
		logger:               logger,
	}
}

// Override Start method to apply authentication
func (a *AuthStreamableHTTPServer) Start(addr string) error {
	a.mu.Lock()
	mux := http.NewServeMux()
	mux.Handle(a.endpointPath, a)
	mux.HandleFunc("/health", a.healthzHandler)

	a.httpServer = &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 300 * time.Second,
	}
	a.mu.Unlock()
	return a.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (a *AuthStreamableHTTPServer) Shutdown(ctx context.Context) error {
	a.mu.RLock()
	srv := a.httpServer
	a.mu.RUnlock()
	if srv != nil {
		return srv.Shutdown(ctx)
	}
	return nil
}

// ServeHTTP implements the http.Handler interface with authentication
func (a *AuthStreamableHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.accessLogging(r)
	token := helper.ExtractBearerToken(r)

	if token == "" {
		jsonRPCError := riskenmcp.NewJSONRPCError(nil, riskenmcp.JSONRPCErrorUnauthorized, "Unauthorized(no authorization header)")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	requestID, err := riskenmcp.ParseJSONRPCRequestID(r)
	if err != nil {
		jsonRPCError := riskenmcp.NewJSONRPCError(nil, riskenmcp.JSONRPCErrorParseError, "Parse error(requestID)")
		http.Error(w, jsonRPCError.String(), http.StatusBadRequest)
		return
	}

	// Verify token
	riskenClient, err := helper.CreateAndValidateRISKENClient(r.Context(), a.riskenURL, token)
	if err != nil {
		a.logger.Error("Failed to validate RISKEN client", slog.String("error", err.Error()))
		jsonRPCError := riskenmcp.NewJSONRPCError(requestID, riskenmcp.JSONRPCErrorUnauthorized, fmt.Sprintf("Invalid RISKEN token: %s", err))
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add RISKEN Client to the request context
	ctx := riskenmcp.WithRISKENClient(r.Context(), riskenClient)
	r = r.WithContext(ctx)

	// Log authenticated request
	a.logger.Debug("Authenticated request",
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	// Delegate to the original handler
	a.StreamableHTTPServer.ServeHTTP(w, r)
}
