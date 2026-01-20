package riskenmcp

import (
	"log/slog"

	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	MCPServer    *server.MCPServer
	riskenClient *risken.Client
	orgClient    *helper.OrganizationClient
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

// NewOrganizationServer creates a new MCP server for Organization Token authentication.
func NewOrganizationServer(orgClient *helper.OrganizationClient, name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	// Create a new MCP server
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)
	mcpserver := createOrganizationMCPServer(s, orgClient, logger)
	return mcpserver
}

// NewOrganizationServerForMultiOrg creates a new MCP server for dynamic Organization Token authentication.
func NewOrganizationServerForMultiOrg(name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	// Create a new MCP server
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)
	mcpserver := createOrganizationMCPServer(
		s,
		nil, // dynamic generate Organization client per request
		logger,
	)
	return mcpserver
}

// NewUnifiedServer creates a new MCP server that supports both Project and Organization tokens.
// Tools are dynamically available based on the token type provided in each request.
func NewUnifiedServer(name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	// Create a new MCP server
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)
	mcpserver := createUnifiedMCPServer(s, logger)
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
	s.AddTool(mcpserver.GetContext())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	s.AddTool(mcpserver.SearchAlert())
	return mcpserver
}

func createOrganizationMCPServer(s *server.MCPServer, orgClient *helper.OrganizationClient, logger *slog.Logger) *Server {
	mcpserver := &Server{
		MCPServer: s,
		orgClient: orgClient,
		logger:    logger,
	}
	s.AddTool(mcpserver.GetContext())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	return mcpserver
}

func createUnifiedMCPServer(s *server.MCPServer, logger *slog.Logger) *Server {
	mcpserver := &Server{
		MCPServer: s,
		logger:    logger,
	}
	// Unified tools (support both Project and Organization tokens)
	s.AddResourceTemplate(mcpserver.GetFindingResource())
	s.AddTool(mcpserver.GetContext())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	s.AddTool(mcpserver.SearchAlert())
	return mcpserver
}
