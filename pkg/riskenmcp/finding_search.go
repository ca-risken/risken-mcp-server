package riskenmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type SearchFindingResponse struct {
	Findings []*finding.Finding `json:"findings,omitempty"`
	Total    uint32             `json:"total"`
	Offset   int32              `json:"offset"`
	Limit    int32              `json:"limit"`
}

func (s *Server) SearchFinding() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_finding",
			mcp.WithDescription("Search RISKEN findings. Use this when a request include \"finding\", \"issue\", \"ファインディング\", \"問題\"..."),
			mcp.WithNumber(
				"finding_id",
				mcp.Description("Finding ID."),
			),
			mcp.WithNumber(
				"alert_id",
				mcp.Description("Alert ID."),
			),
			mcp.WithArray(
				"data_source",
				mcp.Description("RISKEN DataSource. e.g. aws, google, code (like github, gitlab, etc.), osint, diagnosis, azure, ..."),
				mcp.Enum("aws", "google", "code", "osint", "diagnosis", "azure"),
			),
			mcp.WithArray(
				"resource_name",
				mcp.Description("RISKEN ResourceName. e.g. \"arn:aws:iam::123456789012:user/test-user\" ..."),
			),
			mcp.WithNumber(
				"from_score",
				mcp.Description("Minimum score of the findings."),
				mcp.DefaultNumber(0.5),
				mcp.Max(1.0),
				mcp.Min(0.0),
			),
			mcp.WithNumber(
				"status",
				mcp.Description("Status of the findings. (0: all, 1: active, 2: pending)"),
				mcp.DefaultNumber(1),
				mcp.Enum("0", "1", "2"),
			),
			mcp.WithNumber(
				"offset",
				mcp.Description("Offset of the findings."),
				mcp.DefaultNumber(0),
			),
			mcp.WithNumber(
				"limit",
				mcp.Description("Limit of the findings."),
				mcp.DefaultNumber(10),
				mcp.Max(100),
				mcp.Min(1),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			riskenClient, err := s.GetRISKENClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get RISKEN client: %w", err)
			}

			// Parse params
			params, err := s.ParseSearchFindingParams(ctx, req, riskenClient)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to parse params: %s", err)), nil
			}

			// Call RISKEN API
			findings, err := riskenClient.ListFinding(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get findings: %s", err)), nil
			}

			searchResult := &SearchFindingResponse{
				Findings: []*finding.Finding{},
				Total:    uint32(findings.Total),
				Offset:   int32(params.Offset),
				Limit:    int32(params.Limit),
			}
			for _, fid := range findings.FindingId {
				finding, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
					ProjectId: params.ProjectId,
					FindingId: fid,
				})
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to get finding: %s", err)), nil
				}
				searchResult.Findings = append(searchResult.Findings, finding.Finding)
			}
			jsonData, err := json.Marshal(searchResult)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal search result: %s", err)), nil
			}
			return mcp.NewToolResultText(string(jsonData)), nil
		}
}

func (s *Server) ParseSearchFindingParams(ctx context.Context, req mcp.CallToolRequest, riskenClient *risken.Client) (*finding.ListFindingRequest, error) {
	p, err := s.GetCurrentProject(ctx, riskenClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %s", err)
	}
	param := &finding.ListFindingRequest{
		ProjectId: p.ProjectId,
		// Default params
		Offset:    0,
		Limit:     10,
		FromScore: 0.1,
		Status:    finding.FindingStatus_FINDING_ACTIVE,
	}

	// DEBUG
	for k, v := range req.GetArguments() {
		s.logger.Debug("SearchFinding args", slog.String("key", k), slog.Any("value", v), slog.String("type", fmt.Sprintf("%T", v)))
	}

	findingID, err := helper.ParseMCPArgs[float64]("finding_id", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("finding_id error: %s", err)
	}
	if findingID != nil {
		param.FindingId = uint64(*findingID)
		param.FromScore = 0.0
		param.Status = finding.FindingStatus_FINDING_UNKNOWN
		return param, nil // finding_id is specified, so return immediately
	}

	alertID, err := helper.ParseMCPArgs[float64]("alert_id", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("alert_id error: %s", err)
	}
	if alertID != nil {
		param.AlertId = uint32(*alertID)
		param.FromScore = 0.0
		return param, nil // alert_id is specified, so return immediately
	}

	dataSource, err := helper.ParseMCPArgs[[]any]("data_source", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("data_source error: %s", err)
	}
	if dataSource != nil {
		for _, v := range *dataSource {
			param.DataSource = append(param.DataSource, fmt.Sprintf("%v", v))
		}
	}
	resourceName, err := helper.ParseMCPArgs[[]any]("resource_name", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("resource_name error: %s", err)
	}
	if resourceName != nil {
		for _, v := range *resourceName {
			param.ResourceName = append(param.ResourceName, fmt.Sprintf("%v", v))
		}
	}
	fromScore, err := helper.ParseMCPArgs[float64]("from_score", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("from_score error: %s", err)
	}
	if fromScore != nil {
		param.FromScore = float32(*fromScore)
	}
	status, err := helper.ParseMCPArgs[float64]("status", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("status error: %s", err)
	}
	if status != nil {
		param.Status = finding.FindingStatus(int32(*status))
	}
	offset, err := helper.ParseMCPArgs[float64]("offset", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("offset error: %s", err)
	}
	if offset != nil {
		param.Offset = int32(*offset)
	}
	limit, err := helper.ParseMCPArgs[float64]("limit", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("limit error: %s", err)
	}
	if limit != nil {
		param.Limit = int32(*limit)
	}
	return param, nil
}
