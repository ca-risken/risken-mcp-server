package riskenmcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ca-risken/core/proto/project"
	"github.com/ca-risken/go-risken"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetProject(riskenClient *risken.Client) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_project",
			mcp.WithDescription("Get details of the authenticated RISKEN user. Use this when a request include \"me\", \"my\"..."),
			mcp.WithString("reason",
				mcp.Description("Optional: reason the session was created"),
			),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			resp, err := riskenClient.Signin(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to signin: %s", err)), nil
			}

			project, err := riskenClient.ListProject(ctx, &project.ListProjectRequest{
				ProjectId: resp.ProjectID,
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %s", err)), nil
			}

			r, err := json.Marshal(project)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal project: %w", err)
			}
			return mcp.NewToolResultText(string(r)), nil
		}
}
