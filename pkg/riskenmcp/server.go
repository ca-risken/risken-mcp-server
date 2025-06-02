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
	// Create a new MCP server
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)
	mcpserver := createRISKENMCPServer(s, riskenClient, logger)
	return mcpserver
}

func NewServerForMultiProject(name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	// Create a new MCP server
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)
	mcpserver := createRISKENMCPServer(
		s,
		nil, // dynamic generate RISKEN client per request
		logger,
	)
	return mcpserver
}

func addOpts(opts ...server.ServerOption) []server.ServerOption {
	defaultOpts := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	}
	opts = append(defaultOpts, opts...)
	return opts
}

func createRISKENMCPServer(s *server.MCPServer, riskenClient *risken.Client, logger *slog.Logger) *Server {
	mcpserver := &Server{
		MCPServer:    s,
		riskenClient: riskenClient,
		logger:       logger,
	}
	s.AddResourceTemplate(mcpserver.GetFindingResource())
	s.AddTool(mcpserver.GetProject())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	s.AddTool(mcpserver.SearchAlert())
	return mcpserver
}
