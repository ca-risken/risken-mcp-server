package riskenmcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
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
	token := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if token == "" {
		jsonRPCError := NewJSONRPCError(nil, JSONRPCErrorUnauthorized, "Unauthorized(no authorization header)")
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	requestID, err := ParseJSONRPCRequestID(r)
	if err != nil {
		jsonRPCError := NewJSONRPCError(nil, JSONRPCErrorParseError, "Parse error(requestID)")
		http.Error(w, jsonRPCError.String(), http.StatusBadRequest)
		return
	}

	// Verify token
	riskenClient, err := a.createAndValidateRISKENClient(r.Context(), token)
	if err != nil {
		a.logger.Error("Failed to validate RISKEN client", slog.String("error", err.Error()))
		jsonRPCError := NewJSONRPCError(requestID, JSONRPCErrorUnauthorized, fmt.Sprintf("Invalid RISKEN token: %s", err))
		http.Error(w, jsonRPCError.String(), http.StatusUnauthorized)
		return
	}

	// Add RISKEN Client to the request context
	ctx := WithRISKENClient(r.Context(), riskenClient)
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

// createAndValidateRISKENClient creates a new RISKEN client and validates the token.
func (a *AuthStreamableHTTPServer) createAndValidateRISKENClient(ctx context.Context, token string) (*risken.Client, error) {
	client := risken.NewClient(token, risken.WithAPIEndpoint(a.riskenURL))

	resp, err := client.Signin(ctx) // Signin to validate the token
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}
	if resp == nil || resp.ProjectID == 0 {
		return nil, fmt.Errorf("invalid project: %+v", resp)
	}
	return client, nil
}

func (a *AuthStreamableHTTPServer) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		a.logger.Error("Failed to write healthz response", slog.String("error", err.Error()))
	}
}

func (a *AuthStreamableHTTPServer) accessLogging(r *http.Request) {
	jsonRPC := ""
	bodyBytes, err := helper.ReadAndRestoreRequestBody(r)
	if err == nil {
		jsonRPC = string(bodyBytes)
	}

	a.logger.Debug("Received request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("content_type", r.Header.Get("Content-Type")),
		slog.String("mcp_session_id", r.Header.Get("Mcp-Session-Id")),
		slog.String("json_rpc", jsonRPC),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
	)
}
