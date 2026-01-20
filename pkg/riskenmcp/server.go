package riskenmcp

import (
	"log/slog"

	"github.com/ca-risken/go-risken"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	MCPServer    *server.MCPServer
	riskenClient *risken.Client
	logger       *slog.Logger
}

// NewServer creates a unified MCP server for both Project and Organization tokens.
// For stdio mode, pass a riskenClient. For HTTP mode, pass nil (client created per-request).
func NewServer(riskenClient *risken.Client, name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)

	mcpserver := &Server{
		MCPServer:    s,
		riskenClient: riskenClient,
		logger:       logger,
	}

	// Unified tools (support both Project and Organization tokens)
	s.AddResourceTemplate(mcpserver.GetFindingResource())
	s.AddTool(mcpserver.GetContext())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	s.AddTool(mcpserver.SearchAlert())

	return mcpserver
}

// NewServerForMultiProject creates a MCP server for HTTP mode where client is created per-request.
// Deprecated: Use NewServer(nil, ...) instead.
func NewServerForMultiProject(name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	return NewServer(nil, name, version, logger, opts...)
}

func addOpts(opts ...server.ServerOption) []server.ServerOption {
	defaultOpts := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	}
	opts = append(defaultOpts, opts...)
	return opts
}
