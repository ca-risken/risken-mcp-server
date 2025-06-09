package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

// OAuth21Metadata represents OAuth 2.1 Authorization Server Metadata (RFC8414)
// Minimum required fields for MCP implementation
type OAuth21Metadata struct {
	Issuer                        string   `json:"issuer" validate:"required,url"`
	AuthorizationEndpoint         string   `json:"authorization_endpoint" validate:"required,url"`
	TokenEndpoint                 string   `json:"token_endpoint" validate:"required,url"`
	JWKSURI                       string   `json:"jwks_uri" validate:"required,url"`
	RegistrationEndpoint          string   `json:"registration_endpoint,omitempty"`
	ResponseTypesSupported        []string `json:"response_types_supported,omitempty"`
	GrantTypesSupported           []string `json:"grant_types_supported,omitempty"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported,omitempty"`
	ScopesSupported               []string `json:"scopes_supported,omitempty"`
}

// handleProtectedResourceMetadata serves protected resource metadata (RFC9728)
// This is REQUIRED by MCP specification
func (s *Server) handleProtectedResourceMetadata(w http.ResponseWriter, _ *http.Request) {
	metadata := s.config.GenerateProtectedResourceMetadata()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		http.Error(w, "failed to encode protected resource metadata", http.StatusInternalServerError)
	}
}

// ProtectedResourceMetadata for MCP server (RFC9728)
// REQUIRED by MCP specification section 2.3.1
type ProtectedResourceMetadata struct {
	Resource             string   `json:"resource"`
	AuthorizationServers []string `json:"authorization_servers"`
	ScopesProvided       []string `json:"scopes_provided,omitempty"`
	ScopesRequired       []string `json:"scopes_required,omitempty"`
}

// GenerateProtectedResourceMetadata generates metadata for MCP protected resource
// REQUIRED by MCP specification section 2.3.1
func (c *Config) GenerateProtectedResourceMetadata() *ProtectedResourceMetadata {
	return &ProtectedResourceMetadata{
		Resource:             c.MCPServerURL,
		AuthorizationServers: []string{c.MCPServerURL}, // MCP server acts as proxy to IdP
		ScopesProvided:       []string{"openid"},
		ScopesRequired:       []string{"openid"},
	}
}

// handleAuthorizationServerMetadata provides OAuth2.1 metadata for Third-Party Authorization Flow
func (s *Server) handleAuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Create proxy metadata - override endpoints to point to MCP server for Third-Party Authorization Flow
	proxyMetadata := OAuth21Metadata{
		Issuer:                        s.oauth21Metadata.Issuer,
		JWKSURI:                       s.oauth21Metadata.JWKSURI, // Use IdP's JWKS for JWT validation
		AuthorizationEndpoint:         s.config.MCPServerURL + "/authorize",
		TokenEndpoint:                 s.config.MCPServerURL + "/token",
		RegistrationEndpoint:          s.config.MCPServerURL + "/register",
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code"},
		CodeChallengeMethodsSupported: []string{"S256"}, // PKCE REQUIRED by MCP
		ScopesSupported:               []string{"openid"},
	}

	if err := json.NewEncoder(w).Encode(proxyMetadata); err != nil {
		s.logger.Error("Failed to encode authorization server metadata", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Debug("Served proxied OAuth2.1 authorization server metadata for Third-Party Authorization Flow",
		slog.String("issuer", proxyMetadata.Issuer),
		slog.String("authorization_endpoint", proxyMetadata.AuthorizationEndpoint),
		slog.String("token_endpoint", proxyMetadata.TokenEndpoint),
		slog.String("client_ip", helper.ExtractClientIP(r)))
}

// LoadMetadata loads and validates authorization server metadata from IdP
func (s *Server) LoadMetadata(ctx context.Context) error {
	metadata, err := s.fetchAuthorizationServerMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch authorization server metadata: %w", err)
	}

	// Validate using tags
	if err := validate.Struct(metadata); err != nil {
		return fmt.Errorf("invalid IdP metadata: %w", err)
	}

	s.oauth21Metadata = metadata

	s.logger.Info("Validated authorization server metadata from IdP",
		slog.String("issuer", metadata.Issuer),
		slog.String("authz_endpoint", metadata.AuthorizationEndpoint),
		slog.String("token_endpoint", metadata.TokenEndpoint))

	return nil
}

// fetchAuthorizationServerMetadata retrieves metadata from IdP
func (s *Server) fetchAuthorizationServerMetadata(ctx context.Context) (*OAuth21Metadata, error) {
	discoveryURL := s.config.AuthzMetadataEndpoint
	httpClient := helper.NewHTTPClient(s.logger)

	// Use DoSimpleGET since we need to decode to a specific struct, not map[string]any
	responseBody, err := httpClient.DoSimpleGET(ctx, discoveryURL, "Authz_Metadata_Discovery")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery: %w", err)
	}

	var metadata OAuth21Metadata
	if err := json.Unmarshal(responseBody, &metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	s.logger.Debug("Fetched authorization server metadata from IdP",
		slog.String("issuer", metadata.Issuer),
		slog.String("discovery_url", discoveryURL))

	return &metadata, nil
}
