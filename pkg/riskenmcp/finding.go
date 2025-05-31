package riskenmcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) GetFindingResource() (mcp.ResourceTemplate, server.ResourceTemplateHandlerFunc) {
	return mcp.NewResourceTemplate(
			"finding://{project_id}/{finding_id}",
			"RISKEN Finding",
		),
		s.FindingResourceContentsHandler()
}

func (s *Server) FindingResourceContentsHandler() func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		p, err := s.GetCurrentProject(ctx)
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
		finding, err := s.riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
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
