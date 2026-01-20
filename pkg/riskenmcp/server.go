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

// NewServer creates a unified MCP server for both Project and Organization tokens.
// This is used for HTTP mode where the client is created per-request based on the token.
func NewServer(name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)

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

// NewServerWithClient creates a unified MCP server with a pre-authenticated client.
// This is used for Stdio mode where the client is created at startup.
func NewServerWithClient(unifiedClient *helper.UnifiedClient, name, version string, logger *slog.Logger, opts ...server.ServerOption) *Server {
	opts = addOpts(opts...)
	s := server.NewMCPServer(name, version, opts...)

	mcpserver := &Server{
		MCPServer: s,
		logger:    logger,
	}

	// Set the appropriate client based on token type
	switch unifiedClient.TokenType {
	case helper.TokenTypeOrganization:
		mcpserver.orgClient = unifiedClient.OrgClient
	case helper.TokenTypeProject:
		mcpserver.riskenClient = unifiedClient.ProjectClient
	}

	// Unified tools (support both Project and Organization tokens)
	s.AddResourceTemplate(mcpserver.GetFindingResource())
	s.AddTool(mcpserver.GetContext())
	s.AddTool(mcpserver.SearchFinding())
	s.AddTool(mcpserver.ArchiveFinding())
	s.AddTool(mcpserver.SearchAlert())

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
