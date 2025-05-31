package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ca-risken/go-risken"
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
			if err := runStdioServer(); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(stdioCmd)
}

func runStdioServer() error {
	// Create app context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create RISKEN client
	token := os.Getenv("RISKEN_ACCESS_TOKEN")
	if token == "" {
		return fmt.Errorf("RISKEN_ACCESS_TOKEN not set")
	}
	url := os.Getenv("RISKEN_URL")
	if url == "" {
		return fmt.Errorf("RISKEN_URL not set")
	}
	riskenClient := risken.NewClient(token, risken.WithAPIEndpoint(url))

	// Start server
	mcpserver := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion, stdioLogger)
	errC := make(chan error, 1)
	go func() {
		errC <- server.ServeStdio(mcpserver.MCPServer)
	}()
	stdioLogger.Info("starting RISKEN MCP server...")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		stdioLogger.Info("shutting down server...")
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("error running server: %w", err)
		}
	}
	return nil
}
