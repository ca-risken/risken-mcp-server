package oauth

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// AuthorizeRequest represents OAuth2.1 authorization request parameters
type AuthorizeRequest struct {
	ResponseType        string `json:"response_type" validate:"required,eq=code"`
	ClientID            string `json:"client_id" validate:"required"`
	RedirectURI         string `json:"redirect_uri" validate:"required,url"`
	State               string `json:"state" validate:"required"`
	CodeChallenge       string `json:"code_challenge" validate:"required"`
	CodeChallengeMethod string `json:"code_challenge_method" validate:"required,eq=S256"`
	Scope               string `json:"scope,omitempty"`
}

// ParseAuthorizeRequest parses and validates OAuth2.1 authorization request
func ParseAuthorizeRequest(r *http.Request) (*AuthorizeRequest, error) {
	query := r.URL.Query()

	req := &AuthorizeRequest{
		ResponseType:        query.Get("response_type"),
		ClientID:            query.Get("client_id"),
		RedirectURI:         query.Get("redirect_uri"),
		State:               query.Get("state"),
		CodeChallenge:       query.Get("code_challenge"),
		CodeChallengeMethod: query.Get("code_challenge_method"),
		Scope:               query.Get("scope"),
	}

	// Validate using tags
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return req, nil
}

// handleAuthorize handles OAuth authorization requests from MCP clients
func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	// Parse and validate authorization request
	authReq, err := ParseAuthorizeRequest(r)
	if err != nil {
		s.logger.Error("Invalid authorization request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Info("Authorization request received",
		slog.String("client_id", authReq.ClientID),
		slog.String("redirect_uri", authReq.RedirectURI),
		slog.String("state", authReq.State),
		slog.String("code_challenge_method", authReq.CodeChallengeMethod))

	// Store session data
	sessionData := &SessionData{
		State:         authReq.State,         // Original state from MCP client
		CodeChallenge: authReq.CodeChallenge, // Store code_challenge for later verification with code_verifier
		RedirectURI:   authReq.RedirectURI,   // MCP client's redirect URI
		ClientID:      authReq.ClientID,      // MCP client ID (not used but stored for reference)
		CreatedAt:     time.Now(),
	}
	jwtToken, err := s.sessionManager.Store(sessionData)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Use JWT token as the internal state for Auth0
	internalState := jwtToken

	// Build Auth0 authorization URL
	authzURL, err := url.Parse(s.oauth21Metadata.AuthorizationEndpoint)
	if err != nil {
		http.Error(w, "Invalid authorization endpoint", http.StatusInternalServerError)
		return
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", s.config.ClientID)                          // Use our fixed client ID
	params.Set("redirect_uri", s.config.MCPServerURL+"/oauth/callback") // Callback MCP server
	params.Set("state", internalState)                                  // Our internal state
	params.Set("scope", "openid")

	authzURL.RawQuery = params.Encode()

	s.logger.Info("Redirecting to IdP for authorization",
		slog.String("idp_url", authzURL.String()),
		slog.String("internal_state", internalState))

	// Redirect to IdP for authorization
	http.Redirect(w, r, authzURL.String(), http.StatusFound)
}
