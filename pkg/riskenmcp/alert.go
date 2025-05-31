package riskenmcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ca-risken/core/proto/alert"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) SearchAlert() (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_alert",
			mcp.WithDescription("Search RISKEN alert. Use this when a request include \"alert\", \"アラート\" ..."),
			mcp.WithNumber(
				"status",
				mcp.Description("Status of alert. 1: active(有効なアラート), 2: pending(保留中), 3: deactive(解決済みアラート)"),
				mcp.Enum("1", "2", "3"),
				mcp.DefaultNumber(1),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse params
			params, err := s.ParseSearchAlertParams(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to parse params: %s", err)), nil
			}

			// Call RISKEN API
			resp, err := s.riskenClient.ListAlert(ctx, params)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to search alert: %s", err)), nil
			}
			jsonData, err := json.Marshal(resp)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal response: %s", err)), nil
			}
			return mcp.NewToolResultText(string(jsonData)), nil
		}
}

func (s *Server) ParseSearchAlertParams(ctx context.Context, req mcp.CallToolRequest) (*alert.ListAlertRequest, error) {
	p, err := s.GetCurrentProject(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %s", err)
	}
	param := &alert.ListAlertRequest{
		ProjectId: p.ProjectId,
		Status:    []alert.Status{alert.Status_ACTIVE},
	}

	status, err := helper.ParseMCPArgs[float64]("status", req.GetArguments())
	if err != nil {
		return nil, fmt.Errorf("status error: %s", err)
	}
	if status != nil {
		param.Status = []alert.Status{alert.Status(int32(*status))}
	}

	return param, nil
}
