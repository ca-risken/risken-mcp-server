package main

import (
	"log/slog"
	"os"

	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/ca-risken/risken-mcp-server/pkg/streamablehttp"
	"github.com/spf13/cobra"
)

const (
	mcpEndpointPath = "/mcp"
)

var (
	httpPort string

	httpCmd = &cobra.Command{
		Use:   "http",
		Short: "Start Streamable-HTTP MCP server",
		Long:  `Start a server that communicates via Streamable-HTTP.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runHTTPServer()
		},
	}
)

func init() {
	httpCmd.Flags().StringVarP(&httpPort, "port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(httpCmd)
}

func runHTTPServer() error {
	// Set log level based on debug flag
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	httpLogger := logging.NewHTTPLogger(level)

	// Create RISKEN client
	url := os.Getenv("RISKEN_URL")

	// Create MCP server
	mcpserver := riskenmcp.NewServerForMultiProject(ServerName, ServerVersion, httpLogger)
	httpServer := streamablehttp.NewAuthServer(
		mcpserver.MCPServer,
		url,
		mcpEndpointPath,
		httpLogger,
	)

	addr := ":" + httpPort
	httpLogger.Info(
		"Starting RISKEN MCP HTTP server...",
		slog.String("name", ServerName),
		slog.String("version", ServerVersion),
		slog.String("address", addr),
		slog.String("endpoint", mcpEndpointPath),
	)

	// Start server
	return httpServer.Start(addr)
}
