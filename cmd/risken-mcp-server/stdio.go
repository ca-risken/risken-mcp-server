package main

import (
	"log/slog"

	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var (
	stdioLogger = logging.NewStdioLogger(slog.LevelDebug)

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runStdioServer()
		},
	}
)

func init() {
	rootCmd.AddCommand(stdioCmd)
}

func runStdioServer() error {
	// Create RISKEN client
	riskenClient, err := newRISKENClient()
	if err != nil {
		return err
	}

	// Create and start server
	mcpserver := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion, stdioLogger)
	stdioLogger.Info(
		"Starting RISKEN MCP server...",
		slog.String("name", ServerName),
		slog.String("version", ServerVersion),
	)

	// ServeStdio handles signal handling and error management internally
	return server.ServeStdio(mcpserver.MCPServer)
}
