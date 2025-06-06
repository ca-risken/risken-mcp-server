package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

// OAuth21Metadata represents OAuth 2.1 Authorization Server Metadata (RFC8414)
type OAuth21Metadata struct {
	Issuer                   string   `json:"issuer"`
	AuthorizationEndpoint    string   `json:"authorization_endpoint"`
	TokenEndpoint            string   `json:"token_endpoint"`
	JWKSURI                  string   `json:"jwks_uri"`
	ScopesSupported          []string `json:"scopes_supported,omitempty"`
	ResponseTypesSupported   []string `json:"response_types_supported,omitempty"`
	GrantTypesSupported      []string `json:"grant_types_supported,omitempty"`
	TokenEndpointAuthMethods []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	RevocationEndpoint       string   `json:"revocation_endpoint,omitempty"`
	IntrospectionEndpoint    string   `json:"introspection_endpoint,omitempty"`
	RegistrationEndpoint     string   `json:"registration_endpoint,omitempty"`
	// PKCE Support (REQUIRED by MCP)
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported,omitempty"`
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

// handleAuthorizationServerMetadata provides OAuth2.1 metadata for OIDC IdP
func (s *Server) handleAuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	if s.oauth21Metadata == nil {
		s.logger.Error("Authorization server metadata not initialized")
		http.Error(w, "Server not properly initialized", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if err := json.NewEncoder(w).Encode(s.oauth21Metadata); err != nil {
		s.logger.Error("Failed to encode authorization server metadata", slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Debug("Served cached OAuth2.1 authorization server metadata",
		slog.String("issuer", s.oauth21Metadata.Issuer),
		slog.String("client_ip", helper.ExtractClientIP(r)))
}

// LoadMetadata loads and validates authorization server metadata from IdP
func (s *Server) LoadMetadata(ctx context.Context) error {
	metadata, err := s.fetchAuthorizationServerMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch authorization server metadata: %w", err)
	}

	// Validate required endpoints
	if metadata.Issuer == "" {
		return fmt.Errorf("IdP metadata missing issuer")
	}
	if metadata.AuthorizationEndpoint == "" {
		return fmt.Errorf("IdP metadata missing authorization_endpoint")
	}
	if metadata.TokenEndpoint == "" {
		return fmt.Errorf("IdP metadata missing token_endpoint")
	}
	if metadata.JWKSURI == "" {
		return fmt.Errorf("IdP metadata missing jwks_uri")
	}

	// Add MCP server's DCR endpoint if not provided by IdP
	if metadata.RegistrationEndpoint == "" {
		metadata.RegistrationEndpoint = s.config.MCPServerURL + "/register"
		s.logger.Info("Added MCP server DCR endpoint",
			slog.String("registration_endpoint", metadata.RegistrationEndpoint))
	}
	// Add PKCE support (REQUIRED by MCP)
	if len(metadata.CodeChallengeMethodsSupported) == 0 {
		metadata.CodeChallengeMethodsSupported = []string{"S256"}
		s.logger.Info("Added PKCE support", slog.Any("methods", metadata.CodeChallengeMethodsSupported))
	}
	// Add public client support for MCP
	if !slices.Contains(metadata.TokenEndpointAuthMethods, "none") {
		metadata.TokenEndpointAuthMethods = append(metadata.TokenEndpointAuthMethods, "none")
		s.logger.Info("Added public client support")
	}
	if len(metadata.ScopesSupported) == 0 {
		metadata.ScopesSupported = []string{"openid", "email", "profile"}
		s.logger.Info("Added default scopes support", slog.Any("scopes", metadata.ScopesSupported))
	}
	s.oauth21Metadata = metadata

	s.logger.Info("Loaded authorization server metadata from IdP",
		slog.String("issuer", metadata.Issuer),
		slog.String("authz_endpoint", metadata.AuthorizationEndpoint),
		slog.String("token_endpoint", metadata.TokenEndpoint),
		slog.String("registration_endpoint", metadata.RegistrationEndpoint),
		slog.Any("grant_types", metadata.GrantTypesSupported),
		slog.Any("scopes_supported", metadata.ScopesSupported))

	return nil
}

// fetchAuthorizationServerMetadata retrieves metadata from IdP
func (s *Server) fetchAuthorizationServerMetadata(ctx context.Context) (*OAuth21Metadata, error) {
	discoveryURL := s.config.AuthzMetadataEndpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery failed with status: %d", resp.StatusCode)
	}

	var metadata OAuth21Metadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	s.logger.Debug("Fetched authorization server metadata from IdP",
		slog.String("issuer", metadata.Issuer),
		slog.String("discovery_url", discoveryURL))

	return &metadata, nil
}
