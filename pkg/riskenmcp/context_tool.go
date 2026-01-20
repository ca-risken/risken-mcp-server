package riskenmcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ContextResponse represents the response for get_context tool.
type ContextResponse struct {
	TokenType string `json:"token_type"` // "project" or "organization"

	// Organization Token fields
	OrganizationID   uint32           `json:"organization_id,omitempty"`
	OrganizationName string           `json:"organization_name,omitempty"`
	Projects         []ProjectSummary `json:"projects,omitempty"`

	// Project Token fields
	ProjectID   uint32 `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
}

// ProjectSummary represents a summary of a project.
type ProjectSummary struct {
	ProjectID   uint32 `json:"project_id"`
	ProjectName string `json:"project_name"`
}

// GetContext returns a tool for getting the current authentication context.
func (s *Server) GetContext() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_context",
			mcp.WithDescription(`Get the current authentication context.
Returns Organization info if using Organization token, or Project info if using Project token.
Use this to understand the current scope before searching findings.`),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Organization Client を先に試す
			if orgClient, err := s.GetOrganizationClient(ctx); err == nil {
				return s.getContextForOrg(ctx, orgClient)
			}

			// Project Client にフォールバック
			riskenClient, err := s.GetRISKENClient(ctx)
			if err != nil {
				return mcp.NewToolResultError("no client found"), nil
			}
			return s.getContextForProject(ctx, riskenClient)
		}
}

func (s *Server) getContextForOrg(ctx context.Context, orgClient *helper.OrganizationClient) (*mcp.CallToolResult, error) {
	// Organization 情報を取得
	orgResp, err := orgClient.ListOrganization(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get organization: %s", err)), nil
	}
	if len(orgResp.Organization) == 0 {
		return mcp.NewToolResultError("organization not found"), nil
	}
	org := orgResp.Organization[0]

	// Organization 配下のプロジェクト一覧を取得
	projectsResp, err := orgClient.ListProject(ctx)
	if err != nil {
		s.logger.Warn("failed to list projects", "error", err)
	}

	response := &ContextResponse{
		TokenType:        "organization",
		OrganizationID:   org.OrganizationID,
		OrganizationName: org.Name,
		Projects:         []ProjectSummary{},
	}

	if projectsResp != nil {
		for _, p := range projectsResp.Project {
			response.Projects = append(response.Projects, ProjectSummary{
				ProjectID:   p.ProjectID,
				ProjectName: p.Name,
			})
		}
	}

	jsonData, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *Server) getContextForProject(ctx context.Context, riskenClient *risken.Client) (*mcp.CallToolResult, error) {
	project, err := s.GetCurrentProject(ctx, riskenClient)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %s", err)), nil
	}

	response := &ContextResponse{
		TokenType:   "project",
		ProjectID:   project.ProjectId,
		ProjectName: project.Name,
	}

	jsonData, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonData)), nil
}
