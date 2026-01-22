package riskenmcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ca-risken/core/proto/organization"
	"github.com/ca-risken/core/proto/project"
	"github.com/ca-risken/go-risken"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ContextResponse represents the response for get_context tool.
type ContextResponse struct {
	// Organization Token fields
	OrganizationID   uint32 `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`

	// Project Token fields
	ProjectID   uint32 `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
}

// GetContext returns a tool for getting the current authentication context.
func (s *Server) GetContext() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_context",
			mcp.WithDescription(`Get the current authentication context.
Returns Organization info if using Organization token, or Project info if using Project token.`),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			riskenClient, err := s.GetRISKENClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get RISKEN client: %w", err)
			}

			// Determine token type via Signin
			signinResp, err := riskenClient.Signin(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to signin: %s", err)), nil
			}

			// organization_id の有無で分岐
			if signinResp.OrganizationID > 0 {
				return s.getContextForOrg(ctx, riskenClient, signinResp)
			}
			return s.getContextForProject(ctx, riskenClient, signinResp)
		}
}

func (s *Server) getContextForOrg(ctx context.Context, client *risken.Client, signin *risken.SigninResponse) (*mcp.CallToolResult, error) {
	// ListOrganization を呼び出す
	orgResp, err := client.ListOrganization(ctx, &organization.ListOrganizationRequest{
		OrganizationId: signin.OrganizationID,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get organization: %s", err)), nil
	}
	if len(orgResp.Organization) == 0 {
		return mcp.NewToolResultError("organization not found"), nil
	}
	org := orgResp.Organization[0]

	response := &ContextResponse{
		OrganizationID:   org.OrganizationId,
		OrganizationName: org.Name,
	}

	jsonData, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *Server) getContextForProject(ctx context.Context, client *risken.Client, signin *risken.SigninResponse) (*mcp.CallToolResult, error) {
	// ListProject を呼び出す
	projectResp, err := client.ListProject(ctx, &project.ListProjectRequest{
		ProjectId: signin.ProjectID,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %s", err)), nil
	}
	if len(projectResp.Project) == 0 {
		return mcp.NewToolResultError("project not found"), nil
	}
	p := projectResp.Project[0]

	response := &ContextResponse{
		ProjectID:   p.ProjectId,
		ProjectName: p.Name,
	}

	jsonData, _ := json.Marshal(response)
	return mcp.NewToolResultText(string(jsonData)), nil
}
