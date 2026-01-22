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
				return nil, fmt.Errorf("failed to get RISKEN client: %w", err)
			}

			// Parse params
			params, err := s.parseArchiveFindingParams(ctx, req, riskenClient)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to parse params: %s", err)), nil
			}

			// Call RISKEN API
			resp, err := riskenClient.PutPendFinding(ctx, params)
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

func (s *Server) parseArchiveFindingParams(ctx context.Context, req mcp.CallToolRequest, riskenClient *risken.Client) (*finding.PutPendFindingRequest, error) {
	// Determine token type via Signin
	signinResp, err := riskenClient.Signin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %s", err)
	}

	findingID, err := helper.ParseMCPArgs[float64]("finding_id", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("finding_id error: %s", err)
	}
	if findingID == nil {
		return nil, fmt.Errorf("finding_id is required")
	}

	// Get project_id from finding
	projectID, err := s.getProjectIDFromFinding(ctx, riskenClient, signinResp, uint64(*findingID))
	if err != nil {
		return nil, fmt.Errorf("failed to get finding: %s", err)
	}

	// Build request
	param := &finding.PutPendFindingRequest{
		ProjectId: projectID,
		PendFinding: &finding.PendFindingForUpsert{
			ProjectId: projectID,
			FindingId: uint64(*findingID),
			ExpiredAt: time.Now().Add(time.Hour * 24 * 365 * 100).Unix(),
		},
	}

	note, err := helper.ParseMCPArgs[string]("note", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("note error: %s", err)
	}
	if note != nil {
		param.PendFinding.Note = *note
	}

	if param.PendFinding.Note == "" {
		param.PendFinding.Note = "Archived by MCP"
	} else {
		param.PendFinding.Note = fmt.Sprintf("Archived by MCP: %s", *note)
	}

	return param, nil
}

// getProjectIDFromFinding gets the project_id from the finding.
func (s *Server) getProjectIDFromFinding(ctx context.Context, riskenClient *risken.Client, signinResp *risken.SigninResponse, findingID uint64) (uint32, error) {
	if signinResp.OrganizationID > 0 {
		// Organization token: use ListFindingForOrg
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

	// Project token: use GetFinding
	resp, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
		ProjectId: signinResp.ProjectID,
		FindingId: findingID,
	})
	if err != nil {
		return 0, err
	}
	return resp.Finding.ProjectId, nil
}
