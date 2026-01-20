package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
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

	// Auto-detect token type and create client
	unifiedClient, err := helper.DetectAndCreateClient(context.Background(), url, token)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create unified server with the client
	mcpserver := riskenmcp.NewServerWithClient(unifiedClient, ServerName, ServerVersion, stdioLogger)

	// Log startup info with token type
	switch unifiedClient.TokenType {
	case helper.TokenTypeOrganization:
		stdioLogger.Info(
			"Starting RISKEN MCP server...",
			slog.String("name", ServerName),
			slog.String("version", ServerVersion),
			slog.String("token_type", "organization"),
			slog.Uint64("organization_id", uint64(unifiedClient.OrgClient.OrganizationID)),
		)
	case helper.TokenTypeProject:
		stdioLogger.Info(
			"Starting RISKEN MCP server...",
			slog.String("name", ServerName),
			slog.String("version", ServerVersion),
			slog.String("token_type", "project"),
		)
	}

	// ServeStdio handles signal handling and error management internally
	return server.ServeStdio(mcpserver.MCPServer)
}
