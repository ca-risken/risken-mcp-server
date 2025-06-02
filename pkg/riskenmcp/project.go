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

func (s *Server) GetProject() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_project",
			mcp.WithDescription("Get details of the authenticated RISKEN project. Use this when a request include \"project\", \"my project\", \"プロジェクト\"..."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := s.GetCurrentProject(ctx, nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %s", err)), nil
			}

			r, err := json.Marshal(p)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal project: %w", err)
			}
			return mcp.NewToolResultText(string(r)), nil
		}
}

func (s *Server) GetCurrentProject(ctx context.Context, riskenClient *risken.Client) (*project.Project, error) {
	if riskenClient == nil {
		client, err := s.GetRISKENClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get RISKEN client: %w", err)
		}
		riskenClient = client
	}

	resp, err := riskenClient.Signin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}

	project, err := riskenClient.ListProject(ctx, &project.ListProjectRequest{
		ProjectId: resp.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project.Project[0], nil
}
