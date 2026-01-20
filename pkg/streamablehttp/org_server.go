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

// OrgAuthServer is a wrapper for StreamableHTTPServer with Organization Token authentication
type OrgAuthServer struct {
	*server.StreamableHTTPServer
	riskenURL    string
	endpointPath string
	logger       *slog.Logger
	httpServer   *http.Server
	mu           sync.RWMutex
}

// NewOrgAuthServer creates a new Organization authenticated server instance
func NewOrgAuthServer(mcpServer *server.MCPServer, riskenURL, endpointPath string, logger *slog.Logger) *OrgAuthServer {
	return &OrgAuthServer{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath(endpointPath)),
		endpointPath:         endpointPath,
		riskenURL:            riskenURL,
		logger:               logger,
	}
}

// Override Start method to apply Organization authentication
func (a *OrgAuthServer) Start(addr string) error {
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
func (a *OrgAuthServer) Shutdown(ctx context.Context) error {
	a.mu.RLock()
	srv := a.httpServer
	a.mu.RUnlock()
	if srv != nil {
		return srv.Shutdown(ctx)
	}
	return nil
}

// healthzHandler handles health check requests
func (a *OrgAuthServer) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
