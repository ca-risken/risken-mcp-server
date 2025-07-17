package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ca-risken/risken-mcp-server/pkg/logging"
	"github.com/ca-risken/risken-mcp-server/pkg/oauth"
	"github.com/ca-risken/risken-mcp-server/pkg/riskenmcp"
	"github.com/spf13/cobra"
)

var (
	oauthPort string

	oauthCmd = &cobra.Command{
		Use:   "oauth",
		Short: "Start OAuth2.1 MCP server",
		Long:  `Start a server that communicates via OAuth2.1.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runOAuthServer()
		},
	}
)

func init() {
	oauthCmd.Flags().StringVarP(&oauthPort, "port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(oauthCmd)
}

func runOAuthServer() error {
	// Set log level based on debug flag
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	oauthLogger := logging.NewHTTPLogger(level)

	// Create RISKEN client
	url := os.Getenv("RISKEN_URL")

	// Create MCP server
	mcpserver := riskenmcp.NewServerForMultiProject(ServerName, ServerVersion, oauthLogger)
	oauthServer := oauth.NewServer(
		mcpserver.MCPServer,
		&oauth.Config{
			MCPServerURL:          os.Getenv("MCP_SERVER_URL"),
			AuthzMetadataEndpoint: os.Getenv("AUTHZ_METADATA_ENDPOINT"),
			ClientID:              os.Getenv("CLIENT_ID"),
			ClientSecret:          os.Getenv("CLIENT_SECRET"),
			JWTSigningKey:         os.Getenv("JWT_SIGNING_KEY"),
		},
		url,
		mcpEndpointPath,
		oauthLogger,
	)
	if err := oauthServer.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize OAuth server: %w", err)
	}

	addr := ":" + oauthPort
	oauthLogger.Info(
		"Starting RISKEN MCP OAuth server...",
		slog.String("name", ServerName),
		slog.String("version", ServerVersion),
		slog.String("address", addr),
		slog.String("endpoint", mcpEndpointPath),
	)

	// Start server
	return oauthServer.Start(addr)
}
