package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var (
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
	// Set log level based on debug flag
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	stdioLogger := logging.NewStdioLogger(level)

	// Create RISKEN client
	url := os.Getenv("RISKEN_URL")
	token := os.Getenv("RISKEN_ACCESS_TOKEN")
	riskenClient, err := newRISKENClient(url, token)
	if err != nil {
		return err
	}

	// Create and start server
	mcpserver := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion, stdioLogger)

	// Signin to get token type info for logging
	signinResp, err := riskenClient.Signin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to signin: %w", err)
	}
	if signinResp.OrganizationID > 0 {
		stdioLogger.Info(
			"Starting RISKEN MCP server...",
			slog.String("name", ServerName),
			slog.String("version", ServerVersion),
			slog.String("token_type", "organization"),
			slog.Uint64("organization_id", uint64(signinResp.OrganizationID)),
		)
	} else {
		stdioLogger.Info(
			"Starting RISKEN MCP server...",
			slog.String("name", ServerName),
			slog.String("version", ServerVersion),
			slog.String("token_type", "project"),
			slog.Uint64("project_id", uint64(signinResp.ProjectID)),
		)
	}

	// ServeStdio handles signal handling and error management internally
	return server.ServeStdio(mcpserver.MCPServer)
}

func newRISKENClient(url, token string) (*risken.Client, error) {
	if url == "" {
		return nil, fmt.Errorf("RISKEN_URL not set")
	}
	if token == "" {
		return nil, fmt.Errorf("RISKEN_ACCESS_TOKEN not set")
	}
	riskenClient := risken.NewClient(token, risken.WithAPIEndpoint(url))
	return riskenClient, nil
}
