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

// SearchFindingResponse represents the response for finding search.
type SearchFindingResponse struct {
	Findings []*finding.Finding `json:"findings,omitempty"`
	Total    uint32             `json:"total"`
	Offset   int32              `json:"offset"`
	Limit    int32              `json:"limit"`
}

// SearchFinding returns a tool for searching findings.
// Works with both Project and Organization tokens (no branching needed).
func (s *Server) SearchFinding() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_finding",
			mcp.WithDescription("Search RISKEN findings."),
			mcp.WithNumber("finding_id", mcp.Description("Finding ID.")),
			mcp.WithNumber("project_id", mcp.Description("Project ID (optional filter).")),
			mcp.WithNumber("alert_id", mcp.Description("Alert ID.")),
			mcp.WithArray("data_source",
				mcp.Description("RISKEN DataSource."),
				mcp.Enum("aws", "google", "code", "osint", "diagnosis", "azure"),
			),
			mcp.WithArray("resource_name", mcp.Description("RISKEN ResourceName.")),
			mcp.WithNumber("from_score",
				mcp.Description("Minimum score of the findings."),
				mcp.DefaultNumber(0.5),
				mcp.Max(1.0),
				mcp.Min(0.0),
			),
			mcp.WithNumber("status",
				mcp.Description("Status of the findings. (0: all, 1: active, 2: pending)"),
				mcp.DefaultNumber(1),
				mcp.Enum("0", "1", "2"),
			),
			mcp.WithNumber("offset", mcp.Description("Offset."), mcp.DefaultNumber(0)),
			mcp.WithNumber("limit",
				mcp.Description("Limit."),
				mcp.DefaultNumber(10),
				mcp.Max(100),
				mcp.Min(1),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// トークンタイプに関係なく同じ処理（分岐不要）
			riskenClient, err := s.GetRISKENClient(ctx)
			if err != nil {
				return mcp.NewToolResultError("no client found"), nil
			}
			return s.searchFindingImpl(ctx, req, riskenClient)
		}
}

// searchFindingImpl implements the finding search logic.
func (s *Server) searchFindingImpl(ctx context.Context, req mcp.CallToolRequest, riskenClient *risken.Client) (*mcp.CallToolResult, error) {
	params, err := s.parseSearchFindingParams(ctx, req, riskenClient)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse params: %s", err)), nil
	}

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
		f, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
			ProjectId: params.ProjectId,
			FindingId: fid,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get finding: %s", err)), nil
		}
		searchResult.Findings = append(searchResult.Findings, f.Finding)
	}

	jsonData, _ := json.Marshal(searchResult)
	return mcp.NewToolResultText(string(jsonData)), nil
}

// parseSearchFindingParams parses the search finding parameters.
func (s *Server) parseSearchFindingParams(ctx context.Context, req mcp.CallToolRequest, riskenClient *risken.Client) (*finding.ListFindingRequest, error) {
	// Signin してプロジェクト情報を取得
	signinResp, err := riskenClient.Signin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %s", err)
	}

	param := &finding.ListFindingRequest{
		ProjectId: signinResp.ProjectID,
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

	// project_id が指定されている場合は上書き
	projectID, err := helper.ParseMCPArgs[float64]("project_id", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("project_id error: %s", err)
	}
	if projectID != nil {
		param.ProjectId = uint32(*projectID)
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
