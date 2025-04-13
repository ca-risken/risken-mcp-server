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

func ArchiveFinding(riskenClient *risken.Client) (tool mcp.Tool, handler server.ToolHandlerFunc) {
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
			// Parse params
			params, err := ParseArchiveFindingParams(ctx, riskenClient, req)
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

func ParseArchiveFindingParams(ctx context.Context, riskenClient *risken.Client, req mcp.CallToolRequest) (*finding.PutPendFindingRequest, error) {
	p, err := GetCurrentProject(ctx, riskenClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %s", err)
	}
	param := &finding.PutPendFindingRequest{
		ProjectId: p.ProjectId,
		PendFinding: &finding.PendFindingForUpsert{
			ProjectId: p.ProjectId,
			ExpiredAt: time.Now().Add(time.Hour * 24 * 365 * 100).Unix(),
		},
	}

	findingID, err := helper.ParseMCPArgs[float64]("finding_id", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("finding_id error: %s", err)
	}
	if findingID != nil {
		param.PendFinding.FindingId = uint64(*findingID)
	}
	note, err := helper.ParseMCPArgs[string]("note", req.Params.Arguments)
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
