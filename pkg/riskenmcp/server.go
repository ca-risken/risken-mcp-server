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

func NewServer(riskenClient *risken.Client, name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	// Add default options
	defaultOpts := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	}
	opts = append(defaultOpts, opts...)

	// Create a new MCP server
	s := server.NewMCPServer(
		name,
		version,
		opts...,
	)
	mcpserver := &Server{
		MCPServer:    s,
		riskenClient: riskenClient,
		logger:       logger,
	}

	// Add resources
	s.AddResourceTemplate(mcpserver.GetFindingResource())

	// Add tools
	s.AddTool(mcpserver.GetProject())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	s.AddTool(mcpserver.SearchAlert())

	return mcpserver
}
