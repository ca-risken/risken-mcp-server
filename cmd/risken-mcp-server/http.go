package main

import (
	"log/slog"

	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
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
	riskenClient, err := newRISKENClient()
	if err != nil {
		return err
	}

	// Create MCP server
	mcpserver := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion, httpLogger)

	// Create Streamable HTTP server
	httpServer := server.NewStreamableHTTPServer(mcpserver.MCPServer)

	addr := ":" + httpPort
	httpLogger.Info("starting RISKEN MCP HTTP server", "address", addr, "endpoint", "/mcp")

	// Start server (this handles everything internally including signal handling)
	return httpServer.Start(addr)
}
