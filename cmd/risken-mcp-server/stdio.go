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

	// Get environment variables
	url := os.Getenv("RISKEN_URL")
	token := os.Getenv("RISKEN_ACCESS_TOKEN") // Accepts both Project and Organization tokens

	if url == "" {
		return fmt.Errorf("RISKEN_URL not set")
	}
	if token == "" {
		return fmt.Errorf("RISKEN_ACCESS_TOKEN not set")
	}

	// Create single RISKEN client (works with both Project and Organization tokens)
	riskenClient := risken.NewClient(token, risken.WithAPIEndpoint(url))

	// Signin to validate token and get token type info
	signinResp, err := riskenClient.Signin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to signin: %w", err)
	}

	// Log startup info with token type
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

	// Create unified server with the client
	mcpserver := riskenmcp.NewServer(riskenClient, ServerName, ServerVersion, stdioLogger)

	// ServeStdio handles signal handling and error management internally
	return server.ServeStdio(mcpserver.MCPServer)
}
