package riskenmcp

import (
	"github.com/ca-risken/go-risken"
	"github.com/mark3labs/mcp-go/server"
)

func NewServer(riskenClient *risken.Client, name, version string, opts ...server.ServerOption) *server.MCPServer {
	// Add default options
	defaultOpts := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	}
	opts = append(defaultOpts, opts...)

	// Create a new MCP server
	s := server.NewMCPServer(
		name,
		version,
		opts...,
	)

	// Add resources
	s.AddResourceTemplate(GetFindingResource(riskenClient))

	// Add tools
	s.AddTool(GetProject(riskenClient))
	s.AddTool(SearchFinding(riskenClient))
	s.AddTool(ArchiveFinding(riskenClient))
	s.AddTool(SearchAlert(riskenClient))

	return s
}
