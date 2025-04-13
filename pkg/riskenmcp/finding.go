package riskenmcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetFindingResource(riskenClient *risken.Client) (mcp.ResourceTemplate, server.ResourceTemplateHandlerFunc) {
	return mcp.NewResourceTemplate(
			"finding://{project_id}/{finding_id}",
			"RISKEN Finding",
		),
		FindingResourceContentsHandler(riskenClient)
}

func FindingResourceContentsHandler(riskenClient *risken.Client) func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		p, err := GetCurrentProject(ctx, riskenClient)
		if err != nil {
			return nil, errors.New("failed to get project")
		}
		findingID, err := helper.ParseMCPArgs[uint64]("finding_id", request.Params.Arguments)
		if err != nil {
			return nil, errors.New("failed to parse finding_id")
		}
		if findingID == nil {
			return nil, errors.New("finding_id is required")
		}

		// Call RISKEN API
		finding, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
			ProjectId: p.ProjectId,
			FindingId: *findingID,
		})
		if err != nil {
			return nil, errors.New("failed to get finding")
		}
		jsonData, err := json.Marshal(finding)
		if err != nil {
			return nil, errors.New("failed to marshal finding")
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	}
}

func SearchFinding(riskenClient *risken.Client) (tool mcp.Tool, handler server.ToolHandlerFunc) {
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
				mcp.Description("RISKEN DataSource. e.g. aws, google, code, osint, diagnosis, azure, ..."),
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
			// Parse params
			params, err := ParseFindingParams(ctx, riskenClient, req)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to parse params: %s", err)), nil
			}
			// Call RISKEN API
			findings, err := riskenClient.ListFinding(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to get findings: %s", err)), nil
			}
			searchResult := []*finding.Finding{}
			for _, fid := range findings.FindingId {
				finding, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
					ProjectId: params.ProjectId,
					FindingId: fid,
				})
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to get finding: %s", err)), nil
				}
				searchResult = append(searchResult, finding.Finding)
			}
			jsonData, err := json.Marshal(searchResult)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal search result: %s", err)), nil
			}
			return mcp.NewToolResultText(string(jsonData)), nil
		}
}

func ParseFindingParams(ctx context.Context, riskenClient *risken.Client, req mcp.CallToolRequest) (*finding.ListFindingRequest, error) {
	p, err := GetCurrentProject(ctx, riskenClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %s", err)
	}
	param := &finding.ListFindingRequest{
		ProjectId: p.ProjectId,
		// Default params
		Offset:    0,
		Limit:     10,
		FromScore: 0.5,
		Status:    finding.FindingStatus_FINDING_ACTIVE,
	}

	// DEBUG
	// for k, v := range req.Params.Arguments {
	// 	logging.Logger.Info("SearchFinding args", slog.String("key", k), slog.Any("value", v), slog.String("type", fmt.Sprintf("%T", v)))
	// }

	findingID, err := helper.ParseMCPArgs[float64]("finding_id", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("finding_id error: %s", err)
	}
	if findingID != nil {
		param.FindingId = uint64(*findingID)
		param.FromScore = 0.0
		param.Status = finding.FindingStatus_FINDING_UNKNOWN
		return param, nil // finding_id is specified, so return immediately
	}

	alertID, err := helper.ParseMCPArgs[float64]("alert_id", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("alert_id error: %s", err)
	}
	if alertID != nil {
		param.AlertId = uint32(*alertID)
		param.FromScore = 0.0
		return param, nil // alert_id is specified, so return immediately
	}

	dataSource, err := helper.ParseMCPArgs[[]string]("data_source", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("data_source error: %s", err)
	}
	if dataSource != nil {
		param.DataSource = *dataSource
	}
	resourceName, err := helper.ParseMCPArgs[[]string]("resource_name", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("resource_name error: %s", err)
	}
	if resourceName != nil {
		param.ResourceName = *resourceName
	}
	fromScore, err := helper.ParseMCPArgs[float64]("from_score", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("from_score error: %s", err)
	}
	if fromScore != nil {
		param.FromScore = float32(*fromScore)
	}
	status, err := helper.ParseMCPArgs[float64]("status", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("status error: %s", err)
	}
	if status != nil {
		param.Status = finding.FindingStatus(int32(*status))
	}
	offset, err := helper.ParseMCPArgs[float64]("offset", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("offset error: %s", err)
	}
	if offset != nil {
		param.Offset = int32(*offset)
	}
	limit, err := helper.ParseMCPArgs[float64]("limit", req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("limit error: %s", err)
	}
	if limit != nil {
		param.Limit = int32(*limit)
	}
	return param, nil
}
