package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	userAgent   = "risken-mcp-server/0.1.0"
	contentType = "application/json"
)

// OrganizationSigninResponse represents the signin response for Organization Token.
type OrganizationSigninResponse struct {
	OrgAccessTokenID uint32 `json:"org_access_token_id,omitempty"`
	OrganizationID   uint32 `json:"organization_id,omitempty"`
}

// OrganizationClient represents a client for Organization API.
type OrganizationClient struct {
	apiEndpoint string
	apiToken    string
	httpClient  *http.Client

	OrganizationID   uint32
	OrgAccessTokenID uint32
}

// NewOrganizationClient creates a new Organization client.
func NewOrganizationClient(apiEndpoint, apiToken string) *OrganizationClient {
	return &OrganizationClient{
		apiEndpoint: apiEndpoint,
		apiToken:    apiToken,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Signin validates the Organization Token and returns the organization info.
func (c *OrganizationClient) Signin(ctx context.Context) (*OrganizationSigninResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiEndpoint+"/api/v1/signin", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signin failed with status: %d", resp.StatusCode)
	}

	var signinResp OrganizationSigninResponse
	if err := json.NewDecoder(resp.Body).Decode(&signinResp); err != nil {
		return nil, fmt.Errorf("failed to decode signin response: %w", err)
	}

	c.OrganizationID = signinResp.OrganizationID
	c.OrgAccessTokenID = signinResp.OrgAccessTokenID

	return &signinResp, nil
}

// CreateAndValidateOrganizationClient creates and validates an Organization client.
func CreateAndValidateOrganizationClient(ctx context.Context, riskenURL, token string) (*OrganizationClient, error) {
	client := NewOrganizationClient(riskenURL, token)

	resp, err := client.Signin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to signin: %w", err)
	}
	if resp == nil || resp.OrganizationID == 0 {
		return nil, fmt.Errorf("invalid organization token: %+v", resp)
	}
	return client, nil
}

func (c *OrganizationClient) setHeaders(req *http.Request) {
	req.Header.Set("Accept", contentType)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", contentType)
}

// Organization represents the organization entity.
type Organization struct {
	OrganizationID uint32 `json:"organization_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// ListOrganizationResponse represents the list organization response.
type ListOrganizationResponse struct {
	Organization []*Organization `json:"organization"`
}

// ListOrganization gets organization information.
func (c *OrganizationClient) ListOrganization(ctx context.Context) (*ListOrganizationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/organization/list-organization?organization_id=%d", c.apiEndpoint, c.OrganizationID),
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list organization failed with status: %d", resp.StatusCode)
	}

	var dataResp struct {
		Data *ListOrganizationResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dataResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return dataResp.Data, nil
}

// Project represents the project entity.
type Project struct {
	ProjectID uint32 `json:"project_id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// ListProjectsInOrganizationResponse represents the response for listing projects in an organization.
type ListProjectsInOrganizationResponse struct {
	Project []*Project `json:"project"`
}

// ListProjectsInOrganizationResponse represents the response for listing projects in an organization.
// Deprecated: Use ListProject instead.
func (c *OrganizationClient) ListProjectsInOrganization(ctx context.Context) (*ListProjectsInOrganizationResponse, error) {
	return c.ListProject(ctx)
}

// ListProject lists all projects in the organization.
func (c *OrganizationClient) ListProject(ctx context.Context) (*ListProjectsInOrganizationResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/organization/list-project?organization_id=%d", c.apiEndpoint, c.OrganizationID),
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list projects failed with status: %d", resp.StatusCode)
	}

	var dataResp struct {
		Data *ListProjectsInOrganizationResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dataResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return dataResp.Data, nil
}

// FindingStatus represents the finding status enum.
type FindingStatus int32

const (
	FindingStatusUnknown FindingStatus = 0
	FindingStatusActive  FindingStatus = 1
	FindingStatusPending FindingStatus = 2
)

// ListFindingForOrgRequest represents the request for listing findings for an organization.
type ListFindingForOrgRequest struct {
	OrganizationID uint32        `json:"organization_id"`
	DataSource     []string      `json:"data_source,omitempty"`
	ResourceName   []string      `json:"resource_name,omitempty"`
	FromScore      float32       `json:"from_score,omitempty"`
	ToScore        float32       `json:"to_score,omitempty"`
	Tag            []string      `json:"tag,omitempty"`
	Sort           string        `json:"sort,omitempty"`
	Direction      string        `json:"direction,omitempty"`
	Offset         int32         `json:"offset,omitempty"`
	Limit          int32         `json:"limit,omitempty"`
	Status         FindingStatus `json:"status,omitempty"`
	FindingID      uint64        `json:"finding_id,omitempty"`
}

// FindingTag represents a finding tag.
type FindingTag struct {
	FindingTagID uint64 `json:"finding_tag_id"`
	FindingID    uint64 `json:"finding_id"`
	ProjectID    uint32 `json:"project_id"`
	Tag          string `json:"tag"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// PendInfo represents pending finding info.
type PendInfo struct {
	Note      string `json:"note,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	Status    string `json:"status,omitempty"`
	ExpiredAt int64  `json:"expired_at,omitempty"`
}

// Finding represents a finding entity.
type Finding struct {
	FindingID    uint64  `json:"finding_id"`
	Description  string  `json:"description"`
	DataSource   string  `json:"data_source"`
	DataSourceID string  `json:"data_source_id"`
	ResourceName string  `json:"resource_name"`
	ProjectID    uint32  `json:"project_id"`
	OriginalScore float32 `json:"original_score"`
	Score        float32 `json:"score"`
	Data         string  `json:"data,omitempty"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
}

// FindingDetail represents a finding with additional details.
type FindingDetail struct {
	Finding     *Finding      `json:"finding"`
	PendInfo    *PendInfo     `json:"pend_info,omitempty"`
	FindingTags []*FindingTag `json:"finding_tags,omitempty"`
}

// ListFindingForOrgResponse represents the response for listing findings for an organization.
type ListFindingForOrgResponse struct {
	Findings []*FindingDetail `json:"findings"`
	Count    uint32           `json:"count"`
	Total    uint32           `json:"total"`
}

// ListFindingForOrg lists findings for an organization.
func (c *OrganizationClient) ListFindingForOrg(ctx context.Context, params *ListFindingForOrgRequest) (*ListFindingForOrgResponse, error) {
	params.OrganizationID = c.OrganizationID

	parsedURL, err := url.Parse(c.apiEndpoint + "/api/v1/finding/list-finding-for-organization")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	parsedURL.RawQuery = structToQueryParams(params).Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list findings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list findings failed with status: %d", resp.StatusCode)
	}

	var dataResp struct {
		Data *ListFindingForOrgResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dataResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return dataResp.Data, nil
}

// PendFindingForUpsert represents a pending finding for upsert.
type PendFindingForUpsert struct {
	ProjectID uint32 `json:"project_id"`
	FindingID uint64 `json:"finding_id"`
	Note      string `json:"note,omitempty"`
	ExpiredAt int64  `json:"expired_at,omitempty"`
}

// PutPendFindingRequest represents the request for putting a pending finding.
type PutPendFindingRequest struct {
	ProjectID   uint32                `json:"project_id"`
	PendFinding *PendFindingForUpsert `json:"pend_finding"`
}

// PendFinding represents a pending finding.
type PendFinding struct {
	FindingID uint64 `json:"finding_id"`
	ProjectID uint32 `json:"project_id"`
	Note      string `json:"note,omitempty"`
	ExpiredAt int64  `json:"expired_at,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// PutPendFindingResponse represents the response for putting a pending finding.
type PutPendFindingResponse struct {
	PendFinding *PendFinding `json:"pend_finding"`
}

// PutPendFinding archives a finding by putting it in pending status.
func (c *OrganizationClient) PutPendFinding(ctx context.Context, params *PutPendFindingRequest) (*PutPendFindingResponse, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.apiEndpoint+"/api/v1/finding/put-pend-finding",
		bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to put pend finding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("put pend finding failed with status: %d", resp.StatusCode)
	}

	var dataResp struct {
		Data *PutPendFindingResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dataResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return dataResp.Data, nil
}

// structToQueryParams converts a struct to URL query parameters.
func structToQueryParams(s interface{}) url.Values {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	params := url.Values{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		jsonTagParts := strings.Split(tag, ",")
		jsonName := jsonTagParts[0]

		fieldValue := v.Field(i)

		switch fieldValue.Kind() {
		case reflect.Slice:
			for j := 0; j < fieldValue.Len(); j++ {
				params.Add(jsonName, fmt.Sprintf("%v", fieldValue.Index(j)))
			}
		case reflect.String, reflect.Int, reflect.Uint32, reflect.Int32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
			if fieldValue.Interface() != reflect.Zero(fieldValue.Type()).Interface() {
				params.Add(jsonName, fmt.Sprintf("%v", fieldValue.Interface()))
			}
		}
	}
	return params
}
