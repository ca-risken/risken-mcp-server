package main

import (
	"log/slog"
	"os"

	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/ca-risken/risken-mcp-server/pkg/streamablehttp"
	"github.com/spf13/cobra"
)

var (
	orgHTTPPort string

	orgHTTPCmd = &cobra.Command{
		Use:   "org-http",
		Short: "Start Streamable-HTTP MCP server for Organization Token",
		Long:  `Start a server that communicates via Streamable-HTTP with Organization Token authentication.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runOrgHTTPServer()
		},
	}
)

func init() {
	orgHTTPCmd.Flags().StringVarP(&orgHTTPPort, "port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(orgHTTPCmd)
}

func runOrgHTTPServer() error {
	// Set log level based on debug flag
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	httpLogger := logging.NewHTTPLogger(level)

	// Create RISKEN URL
	url := os.Getenv("RISKEN_URL")

	// Create MCP server for Organization
	mcpserver := riskenmcp.NewOrganizationServerForMultiOrg(ServerName, ServerVersion, httpLogger)
	httpServer := streamablehttp.NewOrgAuthServer(
		mcpserver.MCPServer,
		url,
		mcpEndpointPath,
		httpLogger,
	)

	addr := ":" + orgHTTPPort
	httpLogger.Info(
		"Starting RISKEN MCP HTTP server (Organization mode)...",
		slog.String("name", ServerName),
		slog.String("version", ServerVersion),
		slog.String("address", addr),
		slog.String("endpoint", mcpEndpointPath),
	)

	// Start server
	return httpServer.Start(addr)
}
