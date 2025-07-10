package oauth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

// CallbackRequest represents OAuth2.1 callback parameters from IdP
type CallbackRequest struct {
	Code             string `json:"code,omitempty" validate:"required_if=Error ''"`
	State            string `json:"state,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// ParseCallbackRequest parses OAuth2.1 callback request from IdP
func ParseCallbackRequest(r *http.Request) (*CallbackRequest, error) {
	query := r.URL.Query()

	req := &CallbackRequest{
		Code:             query.Get("code"),
		State:            query.Get("state"),
		Error:            query.Get("error"),
		ErrorDescription: query.Get("error_description"),
	}

	// Validate using tags
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return req, nil
}

// handleOAuthCallback handles OAuth callback from IdP
func (s *Server) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Parse and validate callback request
	callbackReq, err := ParseCallbackRequest(r)
	if err != nil {
		s.logger.Error("Invalid callback request", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.logger.Info("OAuth callback received from IdP",
		slog.String("state", callbackReq.State),
		slog.String("error", callbackReq.Error))

	// Handle error from IdP
	if callbackReq.Error != "" {
		s.logger.Error("OAuth error from IdP",
			slog.String("error", callbackReq.Error),
			slog.String("error_description", callbackReq.ErrorDescription))
		http.Error(w, fmt.Sprintf("OAuth error: %s - %s", callbackReq.Error, callbackReq.ErrorDescription), http.StatusBadRequest)
		return
	}

	// Retrieve session data from JWT token
	sessionData, exists := s.sessionManager.Get(callbackReq.State)
	if !exists {
		s.logger.Error("Invalid or expired state parameter", slog.String("state", callbackReq.State))
		http.Error(w, "Invalid or expired session", http.StatusBadRequest)
		return
	}

	// Store IdP authorization code for later token exchange in /token endpoint
	sessionData.IDPCode = callbackReq.Code

	// Generate internal JWT authorization code for MCP client
	jwtSessionManager, ok := s.sessionManager.(*JWTSessionManager)
	if !ok {
		http.Error(w, "Session manager type error", http.StatusInternalServerError)
		return
	}

	internalAuthCode, err := jwtSessionManager.GenerateAuthCode(sessionData)
	if err != nil {
		s.logger.Error("Failed to generate authorization code", slog.String("error", err.Error()))
		http.Error(w, "Authorization code generation failed", http.StatusInternalServerError)
		return
	}

	// Redirect back to MCP client with internal authorization code
	clientRedirectURL, err := url.Parse(sessionData.RedirectURI)
	if err != nil {
		http.Error(w, "Invalid client redirect URI", http.StatusInternalServerError)
		return
	}

	params := url.Values{}
	params.Set("code", internalAuthCode) // Use internal JWT authorization code

	// Include state if it was provided by the client (non-empty)
	if sessionData.State != "" {
		params.Set("state", sessionData.State) // Original state from MCP client
	}

	clientRedirectURL.RawQuery = params.Encode()

	// Redirect back to MCP client
	http.Redirect(w, r, clientRedirectURL.String(), http.StatusFound)
}

// exchangeCodeForToken exchanges authorization code for access token with IdP
func (s *Server) exchangeCodeForToken(ctx context.Context, code string) (string, error) {
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("client_id", s.config.ClientID)
	formData.Set("client_secret", s.config.ClientSecret)
	formData.Set("code", code)
	formData.Set("redirect_uri", s.config.MCPServerURL+"/oauth/callback")

	httpClient := helper.NewHTTPClient(s.logger)
	resp, err := httpClient.DoJSONRequestWithValidation(ctx, helper.JSONRequest{
		Method: http.MethodPost,
		URL:    s.oauth21Metadata.TokenEndpoint,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:    formData.Encode(),
		LogName: "Token_Exchange",
	})
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}

	accessToken, ok := resp.Body["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("invalid access_token in response")
	}

	return accessToken, nil
}
