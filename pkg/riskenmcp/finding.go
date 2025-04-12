package riskenmcp

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/go-risken"
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
		// Parse arguments
		projectID, ok := request.Params.Arguments["project_id"].([]string)
		if !ok || len(projectID) == 0 {
			p, err := GetCurrentProject(ctx, riskenClient)
			if err != nil {
				return nil, errors.New("failed to get project")
			}
			projectID = []string{strconv.FormatUint(uint64(p.ProjectId), 10)}
		}
		projectIDUint, err := strconv.ParseUint(projectID[0], 10, 32)
		if err != nil {
			return nil, errors.New("failed to parse project_id")
		}
		findingID, ok := request.Params.Arguments["finding_id"].([]string)
		if !ok || len(findingID) == 0 {
			return nil, errors.New("finding_id is required")
		}
		findingIDUint, err := strconv.ParseUint(findingID[0], 10, 64)
		if err != nil {
			return nil, errors.New("failed to parse finding_id")
		}

		// Call RISKEN API
		finding, err := riskenClient.GetFinding(ctx, &finding.GetFindingRequest{
			ProjectId: uint32(projectIDUint),
			FindingId: findingIDUint,
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
