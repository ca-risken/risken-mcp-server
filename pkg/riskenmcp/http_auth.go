package riskenmcp

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// AuthStreamableHTTPServer is a wrapper for StreamableHTTPServer with authentication
type AuthStreamableHTTPServer struct {
	*server.StreamableHTTPServer
	mcpAuthToken string
	endpointPath string
	logger       *slog.Logger
	httpServer   *http.Server
	mu           sync.RWMutex
}

// NewAutStreamableHTTPServer creates a new authenticated server instance
func NewAuthStreamableHTTPServer(mcpServer *server.MCPServer, mcpAuthToken, endpointPath string, logger *slog.Logger) *AuthStreamableHTTPServer {
	return &AuthStreamableHTTPServer{
		StreamableHTTPServer: server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath(endpointPath)),
		mcpAuthToken:         mcpAuthToken,
		endpointPath:         endpointPath,
		logger:               logger,
	}
}

// Override Start method to apply authentication
func (a *AuthStreamableHTTPServer) Start(addr string) error {
	a.mu.Lock()
	mux := http.NewServeMux()
	mux.Handle(a.endpointPath, a)
	mux.HandleFunc("/healthz", a.healthzHandler)

	a.httpServer = &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 30 * time.Second,
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
	if a.mcpAuthToken != "" {
		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.logger.Warn("Unauthorized request: missing authorization header",
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			a.logger.Warn("Unauthorized request: invalid authorization format",
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			http.Error(w, "Invalid authorization format. Use 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		// Verify token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != a.mcpAuthToken {
			a.logger.Warn("Unauthorized request: invalid MCP auth token",
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
			http.Error(w, "Invalid MCP auth token", http.StatusUnauthorized)
			return
		}

		// Log authenticated request
		a.logger.Debug("Authenticated request",
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
	}

	// Delegate to the original handler
	a.StreamableHTTPServer.ServeHTTP(w, r)
}

func (a *AuthStreamableHTTPServer) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		a.logger.Error("Failed to write healthz response", slog.String("error", err.Error()))
	}
}
