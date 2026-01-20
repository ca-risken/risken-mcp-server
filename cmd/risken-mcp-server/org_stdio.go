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
	orgStdioCmd = &cobra.Command{
		Use:   "org-stdio",
		Short: "Start stdio server for Organization Token",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages with Organization Token authentication.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runOrgStdioServer()
		},
	}
)

func init() {
	rootCmd.AddCommand(orgStdioCmd)
}

func runOrgStdioServer() error {
	// Set log level based on debug flag
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	stdioLogger := logging.NewStdioLogger(level)

	// Create Organization client
	url := os.Getenv("RISKEN_URL")
	token := os.Getenv("RISKEN_ORGANIZATION_TOKEN")
	orgClient, err := newOrganizationClient(url, token)
	if err != nil {
		return err
	}

	// Create and start server
	mcpserver := riskenmcp.NewOrganizationServer(orgClient, ServerName, ServerVersion, stdioLogger)
	stdioLogger.Info(
		"Starting RISKEN MCP server (Organization mode)...",
		slog.String("name", ServerName),
		slog.String("version", ServerVersion),
		slog.Uint64("organization_id", uint64(orgClient.OrganizationID)),
	)

	// ServeStdio handles signal handling and error management internally
	return server.ServeStdio(mcpserver.MCPServer)
}

func newOrganizationClient(url, token string) (*helper.OrganizationClient, error) {
	if url == "" {
		return nil, fmt.Errorf("RISKEN_URL not set")
	}
	if token == "" {
		return nil, fmt.Errorf("RISKEN_ORGANIZATION_TOKEN not set")
	}
	ctx := context.Background()
	orgClient, err := helper.CreateAndValidateOrganizationClient(ctx, url, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization client: %w", err)
	}
	return orgClient, nil
}
