package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

// TokenRequest represents OAuth2.1 token request parameters
type TokenRequest struct {
	GrantType    string `json:"grant_type" validate:"required,eq=authorization_code"` // REQUIRED: "authorization_code"
	Code         string `json:"code" validate:"required"`                             // REQUIRED
	RedirectURI  string `json:"redirect_uri" validate:"required,url"`                 // REQUIRED
	ClientID     string `json:"client_id" validate:"required"`                        // REQUIRED for public clients
	CodeVerifier string `json:"code_verifier" validate:"required,min=43,max=128"`     // REQUIRED for PKCE
}

// ParseTokenRequest parses and validates OAuth2.1 token request
func ParseTokenRequest(r *http.Request) (*TokenRequest, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form data: %w", err)
	}

	req := &TokenRequest{
		GrantType:    r.Form.Get("grant_type"),
		Code:         r.Form.Get("code"),
		RedirectURI:  r.Form.Get("redirect_uri"),
		ClientID:     r.Form.Get("client_id"),
		CodeVerifier: r.Form.Get("code_verifier"),
	}

	// Validate using tags
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return req, nil
}

// TokenResponse represents OAuth2.1 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// NewTokenResponse creates a standard token response
func NewTokenResponse(accessToken string) *TokenResponse {
	return &TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "openid profile email",
	}
}

// handleToken handles token requests from MCP clients
func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate token request
	tokenReq, err := ParseTokenRequest(r)
	if err != nil {
		s.logger.Error("Invalid token request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Info("Token request received",
		slog.String("grant_type", tokenReq.GrantType),
		slog.String("redirect_uri", tokenReq.RedirectURI),
		slog.String("client_id", tokenReq.ClientID))

	// Note: In this simplified implementation, we use the IdP access token directly as the "code"
	// In a proper implementation, you would store the mapping between authorization code and access token
	idpAccessToken := tokenReq.Code

	// TODO: Implement PKCE verification if needed
	// Currently, PKCE is validated between MCP Client ↔ MCP Server
	// but not enforced between MCP Server ↔ IdP
	s.logger.Debug("PKCE verification skipped - IdP doesn't require PKCE",
		slog.String("code_verifier_length", fmt.Sprintf("%d", len(tokenReq.CodeVerifier))))

	tokenResponse := NewTokenResponse(idpAccessToken)

	helper.WriteJSONResponse(w, http.StatusOK, tokenResponse)
	s.logger.Info("Token issued successfully",
		slog.String("redirect_uri", tokenReq.RedirectURI))
}

// VerifyPKCE verifies PKCE code_verifier against code_challenge
func VerifyPKCE(codeChallenge, codeVerifier string) bool {
	expectedChallenge := generateCodeChallenge(codeVerifier)
	return codeChallenge == expectedChallenge
}

// generateCodeChallenge generates PKCE code challenge from verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
