package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ca-risken/go-risken"
)

// TokenType represents the type of RISKEN token.
type TokenType int

const (
	TokenTypeUnknown TokenType = iota
	TokenTypeProject
	TokenTypeOrganization
)

// UnifiedSigninResponse represents a signin response that can be either Project or Organization.
type UnifiedSigninResponse struct {
	// Project Token fields
	ProjectID     uint32 `json:"project_id,omitempty"`
	AccessTokenID uint32 `json:"access_token_id,omitempty"`

	// Organization Token fields
	OrgAccessTokenID uint32 `json:"org_access_token_id,omitempty"`
	OrganizationID   uint32 `json:"organization_id,omitempty"`
}

// TokenType returns the type of the token based on the signin response.
func (r *UnifiedSigninResponse) TokenType() TokenType {
	if r.OrganizationID > 0 {
		return TokenTypeOrganization
	}
	if r.ProjectID > 0 {
		return TokenTypeProject
	}
	return TokenTypeUnknown
}

// UnifiedClient holds either a Project client or Organization client.
type UnifiedClient struct {
	TokenType     TokenType
	ProjectClient *risken.Client
	OrgClient     *OrganizationClient
}

// DetectAndCreateClient detects the token type and creates the appropriate client.
func DetectAndCreateClient(ctx context.Context, riskenURL, token string) (*UnifiedClient, error) {
	// First, try to signin to detect the token type
	signinResp, err := unifiedSignin(ctx, riskenURL, token)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}

	switch signinResp.TokenType() {
	case TokenTypeProject:
		client := risken.NewClient(token, risken.WithAPIEndpoint(riskenURL))
		return &UnifiedClient{
			TokenType:     TokenTypeProject,
			ProjectClient: client,
		}, nil
	case TokenTypeOrganization:
		orgClient := NewOrganizationClient(riskenURL, token)
		orgClient.OrganizationID = signinResp.OrganizationID
		orgClient.OrgAccessTokenID = signinResp.OrgAccessTokenID
		return &UnifiedClient{
			TokenType: TokenTypeOrganization,
			OrgClient: orgClient,
		}, nil
	default:
		return nil, fmt.Errorf("unknown token type: %+v", signinResp)
	}
}

// unifiedSignin performs a signin request and returns the unified response.
func unifiedSignin(ctx context.Context, riskenURL, token string) (*UnifiedSigninResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, riskenURL+"/api/v1/signin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signin failed with status: %d", resp.StatusCode)
	}

	var signinResp UnifiedSigninResponse
	if err := json.NewDecoder(resp.Body).Decode(&signinResp); err != nil {
		return nil, fmt.Errorf("failed to decode signin response: %w", err)
	}

	return &signinResp, nil
}

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
