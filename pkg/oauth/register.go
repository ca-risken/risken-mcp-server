package oauth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

// RegistrationRequest represents OAuth 2.0 Dynamic Client Registration request
type RegistrationRequest struct {
	RedirectURIs            []string `json:"redirect_uris" validate:"required,dive,url"`
	ClientName              string   `json:"client_name,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	ApplicationType         string   `json:"application_type,omitempty"`
}

// RegistrationResponse represents OAuth 2.0 Dynamic Client Registration response
type RegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"` // Only for confidential clients
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name,omitempty"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	ApplicationType         string   `json:"application_type"`
}

// ParseRegistrationRequest parses and validates Dynamic Client Registration request
func ParseRegistrationRequest(r *http.Request) (*RegistrationRequest, error) {
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to parse registration request: %w", err)
	}

	// Validate using tags
	if err := validate.Struct(&req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &req, nil
}

// handleRegister handles OAuth 2.0 Dynamic Client Registration requests
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate registration request
	regReq, err := ParseRegistrationRequest(r)
	if err != nil {
		s.logger.Error("Invalid registration request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Info("Dynamic Client Registration request received",
		slog.String("client_name", regReq.ClientName),
		slog.Any("redirect_uris", regReq.RedirectURIs),
		slog.String("application_type", regReq.ApplicationType))

	// Simplified implementation: Generate or use fixed client credentials
	// In production, you would validate redirect_uris and store client data
	clientID := generateClientID()

	// Set defaults for missing fields
	grantTypes := regReq.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code"}
	}

	responseTypes := regReq.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	authMethod := regReq.TokenEndpointAuthMethod
	if authMethod == "" {
		authMethod = "none" // Public client (no client_secret required)
	}

	applicationType := regReq.ApplicationType
	if applicationType == "" {
		applicationType = "native" // Default for MCP clients
	}

	// Create registration response
	response := &RegistrationResponse{
		ClientID: clientID,
		// ClientSecret is omitted for public clients (MCP typically uses public clients)
		RedirectURIs:            regReq.RedirectURIs,
		ClientName:              regReq.ClientName,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: authMethod,
		ApplicationType:         applicationType,
	}

	helper.WriteJSONResponse(w, http.StatusCreated, response)

	s.logger.Info("Client registered successfully",
		slog.String("client_id", clientID),
		slog.String("client_name", regReq.ClientName),
		slog.String("auth_method", authMethod))
}

// generateClientID generates a unique client ID for the registered client
func generateClientID() string {
	// Simplified implementation: return a fixed client ID
	// In production, you would generate a unique ID per client
	return "mcp-public-client"
}
