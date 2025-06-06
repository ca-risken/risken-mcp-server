package helper

import (
	"context"
	"fmt"

	"github.com/ca-risken/go-risken"
)

// CreateAndValidateRISKENClient creates a new RISKEN client and validates the token.
func CreateAndValidateRISKENClient(ctx context.Context, riskenURL, token string) (*risken.Client, error) {
	client := risken.NewClient(token, risken.WithAPIEndpoint(riskenURL))

	resp, err := client.Signin(ctx) // Signin to validate the token
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}
	if resp == nil || resp.ProjectID == 0 {
		return nil, fmt.Errorf("invalid project: %+v", resp)
	}
	return client, nil
}
