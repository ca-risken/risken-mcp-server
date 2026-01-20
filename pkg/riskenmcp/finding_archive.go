package riskenmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) ArchiveFinding() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("archive_finding",
			mcp.WithDescription("Archive RISKEN finding. Use this when a request include \"archive\", \"アーカイブ\", \"ペンディング\"..."),
			mcp.WithNumber(
				"finding_id",
				mcp.Description("Finding ID."),
				mcp.Required(),
			),
			mcp.WithString(
				"note",
				mcp.Description("Note. ex) This is no risk finding."),
				mcp.DefaultString("Archived by MCP"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			riskenClient, err := s.GetRISKENClient(ctx)
			if err != nil {
				return mcp.NewToolResultError("no client found"), nil
			}

			// Parse finding_id and note
			findingID, err := helper.ParseMCPArgs[float64]("finding_id", req.GetArguments())
			if err != nil || findingID == nil {
				return mcp.NewToolResultError("finding_id is required"), nil
			}

			note := "Archived by MCP"
			if n, _ := helper.ParseMCPArgs[string]("note", req.GetArguments()); n != nil && *n != "" {
				note = fmt.Sprintf("Archived by MCP: %s", *n)
			}

			// Signin でトークンタイプを判定
			signinResp, err := riskenClient.Signin(ctx)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to signin: %s", err)), nil
			}

			// Finding を取得して project_id を特定
			projectID, err := s.getProjectIDFromFinding(ctx, riskenClient, signinResp, uint64(*findingID))
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get finding: %s", err)), nil
			}

			// Archive finding
			resp, err := riskenClient.PutPendFinding(ctx, &finding.PutPendFindingRequest{
				ProjectId: projectID,
				PendFinding: &finding.PendFindingForUpsert{
					ProjectId: projectID,
					FindingId: uint64(*findingID),
					Note:      note,
					ExpiredAt: time.Now().Add(time.Hour * 24 * 365 * 100).Unix(),
				},
			})
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to archive finding: %s", err)), nil
			}
			jsonData, err := json.Marshal(resp)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %s", err)), nil
			}
			return mcp.NewToolResultText(string(jsonData)), nil
		}
}

// getProjectIDFromFinding gets the project_id from the finding.
func (s *Server) getProjectIDFromFinding(ctx context.Context, riskenClient *risken.Client, signinResp *risken.SigninResponse, findingID uint64) (uint32, error) {
	if signinResp.OrganizationID > 0 {
		// Organization トークンの場合は ListFindingForOrg で取得
		resp, err := riskenClient.ListFindingForOrg(ctx, &finding.ListFindingForOrgRequest{
			OrganizationId: signinResp.OrganizationID,
			FindingId:      findingID,
			FromScore:      0.0,
			Status:         finding.FindingStatus_FINDING_UNKNOWN,
			Limit:          1,
		})
		if err != nil {
			return 0, err
		}
		if len(resp.Findings) == 0 {
			return 0, fmt.Errorf("finding not found: %d", findingID)
		}
		return resp.Findings[0].Finding.ProjectId, nil
	}

	// Project トークンの場合は GetFinding で取得
	resp, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
		ProjectId: signinResp.ProjectID,
		FindingId: findingID,
	})
	if err != nil {
		return 0, err
	}
	return resp.Finding.ProjectId, nil
}
