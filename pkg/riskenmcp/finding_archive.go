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

// archiveFindingParams represents the parameters for archiving a finding.
type archiveFindingParams struct {
	ProjectID uint32
	FindingID uint64
	Note      string
}

// ArchiveFinding returns a tool for archiving findings.
// Works with both Project and Organization tokens (no branching needed).
func (s *Server) ArchiveFinding() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("archive_finding",
			mcp.WithDescription("Archive RISKEN finding."),
			mcp.WithNumber("project_id",
				mcp.Description("Project ID that the finding belongs to."),
				mcp.Required(),
			),
			mcp.WithNumber("finding_id",
				mcp.Description("Finding ID."),
				mcp.Required(),
			),
			mcp.WithString("note",
				mcp.Description("Note. ex) This is no risk finding."),
				mcp.DefaultString("Archived by MCP"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// トークンタイプに関係なく同じ処理（分岐不要）
			riskenClient, err := s.GetRISKENClient(ctx)
			if err != nil {
				return mcp.NewToolResultError("no client found"), nil
			}
			return s.archiveFindingImpl(ctx, req, riskenClient)
		}
}

// archiveFindingImpl implements the finding archive logic.
func (s *Server) archiveFindingImpl(ctx context.Context, req mcp.CallToolRequest, riskenClient *risken.Client) (*mcp.CallToolResult, error) {
	params, err := s.parseArchiveFindingParams(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse params: %s", err)), nil
	}

	resp, err := riskenClient.PutPendFinding(ctx, &finding.PutPendFindingRequest{
		ProjectId: params.ProjectID,
		PendFinding: &finding.PendFindingForUpsert{
			ProjectId: params.ProjectID,
			FindingId: params.FindingID,
			Note:      params.Note,
			ExpiredAt: time.Now().Add(time.Hour * 24 * 365 * 100).Unix(),
		},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to archive finding: %s", err)), nil
	}

	jsonData, _ := json.Marshal(resp)
	return mcp.NewToolResultText(string(jsonData)), nil
}

// parseArchiveFindingParams parses the archive finding parameters.
func (s *Server) parseArchiveFindingParams(req mcp.CallToolRequest) (*archiveFindingParams, error) {
	projectID, err := helper.ParseMCPArgs[float64]("project_id", req.GetArguments())
	if err != nil || projectID == nil {
		return nil, fmt.Errorf("project_id is required")
	}

	findingID, err := helper.ParseMCPArgs[float64]("finding_id", req.GetArguments())
	if err != nil || findingID == nil {
		return nil, fmt.Errorf("finding_id is required")
	}

	note := "Archived by MCP"
	if n, _ := helper.ParseMCPArgs[string]("note", req.GetArguments()); n != nil && *n != "" {
		note = fmt.Sprintf("Archived by MCP: %s", *n)
	}

	return &archiveFindingParams{
		ProjectID: uint32(*projectID),
		FindingID: uint64(*findingID),
		Note:      note,
	}, nil
}
