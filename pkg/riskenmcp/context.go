package riskenmcp

import (
	"context"
	"fmt"

	"github.com/ca-risken/core/proto/project"
	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

type contextKey string

const (
	RISKENClientContextKey       contextKey = "risken_client"
	OrganizationClientContextKey contextKey = "organization_client"
)

// WithRISKENClient sets the RISKEN client in the context.
func WithRISKENClient(ctx context.Context, client *risken.Client) context.Context {
	return context.WithValue(ctx, RISKENClientContextKey, client)
}

// GetRISKENClient returns the RISKEN client from server field or context.
func (s *Server) GetRISKENClient(ctx context.Context) (*risken.Client, error) {
	if s.riskenClient != nil {
		return s.riskenClient, nil
	}

	client, ok := ctx.Value(RISKENClientContextKey).(*risken.Client)
	if !ok || client == nil {
		return nil, fmt.Errorf("no RISKEN client found in context")
	}
	return client, nil
}

// WithOrganizationClient sets the Organization client in the context.
func WithOrganizationClient(ctx context.Context, client *helper.OrganizationClient) context.Context {
	return context.WithValue(ctx, OrganizationClientContextKey, client)
}

// GetOrganizationClient returns the Organization client from server field or context.
func (s *Server) GetOrganizationClient(ctx context.Context) (*helper.OrganizationClient, error) {
	if s.orgClient != nil {
		return s.orgClient, nil
	}

	client, ok := ctx.Value(OrganizationClientContextKey).(*helper.OrganizationClient)
	if !ok || client == nil {
		return nil, fmt.Errorf("no Organization client found in context")
	}
	return client, nil
}

// GetCurrentProject returns the current project information.
func (s *Server) GetCurrentProject(ctx context.Context, riskenClient *risken.Client) (*project.Project, error) {
	if riskenClient == nil {
		client, err := s.GetRISKENClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get RISKEN client: %w", err)
		}
		riskenClient = client
	}

	resp, err := riskenClient.Signin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}

	project, err := riskenClient.ListProject(ctx, &project.ListProjectRequest{
		ProjectId: resp.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project.Project[0], nil
}
