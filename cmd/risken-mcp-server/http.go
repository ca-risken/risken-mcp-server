package main

import (
	"log/slog"
	"os"

	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/spf13/cobra"
)

const (
	httpEndpointPath = "/mcp"
)

var (
	httpLogger = logging.NewHTTPLogger(slog.LevelDebug)
	httpPort   string

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
	// Create RISKEN client
	url := os.Getenv("RISKEN_URL")
	token := os.Getenv("RISKEN_ACCESS_TOKEN")
	riskenClient, err := newRISKENClient(url, token)
	if err != nil {
		return err
	}

	// Create MCP server
	mcpserver := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion, httpLogger)
	httpServer := riskenmcp.NewAuthStreamableHTTPServer(
		mcpserver.MCPServer,
		token,
		httpEndpointPath,
		httpLogger,
	)

	addr := ":" + httpPort
	httpLogger.Info(
		"Starting RISKEN MCP HTTP server...",
		slog.String("name", ServerName),
		slog.String("version", ServerVersion),
		slog.String("address", addr),
		slog.Bool("mcp_auth", token != ""),
		slog.String("endpoint", httpEndpointPath),
	)

	// Start server
	return httpServer.Start(addr)
}
