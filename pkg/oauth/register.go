package oauth

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// handleDynamicClientRegistration handles OAuth 2.0 Dynamic Client Registration
func (s *Server) handleDynamicClientRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse client registration request
	var registrationRequest map[string]any
	if err := json.NewDecoder(r.Body).Decode(&registrationRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	s.logger.Info("MCP Client Registration Request",
		slog.Any("redirect_uris", registrationRequest["redirect_uris"]),
		slog.Any("full_request", registrationRequest))

	response := map[string]any{
		"client_id":                  s.config.ClientID,
		"client_secret":              s.config.ClientSecret,
		"token_endpoint_auth_method": "client_secret_post",
		"registration_access_token":  "not-required",
		"registration_client_uri":    s.config.MCPServerURL + "/register",
		"redirect_uris":              registrationRequest["redirect_uris"],
		"grant_types":                []string{"authorization_code", "refresh_token"},
		"response_types":             []string{"code"},
		"scope":                      "openid",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode dynamic client registration response",
			slog.String("error", err.Error()))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	s.logger.Info("Dynamic client registration completed",
		slog.String("client_id", response["client_id"].(string)))
}
