package helper

import (
	"context"
	"fmt"

	"github.com/ca-risken/go-risken"
)

// CreateAndValidateRISKENClient creates a new RISKEN client and validates the token.
// Works with both Project and Organization tokens.
func CreateAndValidateRISKENClient(ctx context.Context, riskenURL, token string) (*risken.Client, error) {
	client := risken.NewClient(token, risken.WithAPIEndpoint(riskenURL))

	resp, err := client.Signin(ctx) // Signin to validate the token
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("invalid signin response")
	}
	// Accept both Project and Organization tokens
	if resp.ProjectID == 0 && resp.OrganizationID == 0 {
		return nil, fmt.Errorf("invalid token: no project_id or organization_id")
	}
	return client, nil
}
