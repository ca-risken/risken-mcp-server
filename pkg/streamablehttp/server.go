package streamablehttp

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
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
