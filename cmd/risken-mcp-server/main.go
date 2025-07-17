package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	ServerName    = "RISKEN MCP Server"
	ServerVersion = "0.0.1"
)

var (
	version = "version"
	commit  = "commit"
	date    = "date"
	debug   bool

	rootCmd = &cobra.Command{
		Use:          "risken-mcp-server",
		Short:        "RISKEN MCP Server",
		Long:         `A RISKEN MCP server that handles various tools and resources.`,
		Version:      fmt.Sprintf("%s %s %s", version, commit, date),
		SilenceUsage: true,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
}

func main() {
	rootCmd.SetOut(os.Stderr)
	rootCmd.SetErr(os.Stderr)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
