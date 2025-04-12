package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

const (
	ServerName    = "RISKEN MCP Server"
	ServerVersion = "0.0.1"
)

var version = "version"
var commit = "commit"
var date = "date"
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

var (
	rootCmd = &cobra.Command{
		Use:     "risken-mcp-server",
		Short:   "RISKEN MCP Server",
		Long:    `A RISKEN MCP server that handles various tools and resources.`,
		Version: fmt.Sprintf("%s %s %s", version, commit, date),
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := runStdioServer(); err != nil {
				logger.Error("failed to run stdio server", slog.Any("error", err))
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
	riskenServer := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion)
	logger.Info("Starting RISKEN MCP server...")
	errC := make(chan error, 1)
	go func() {
		errC <- server.ServeStdio(riskenServer)
	}()

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		logger.Info("shutting down server...")
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("error running server: %w", err)
		}
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
