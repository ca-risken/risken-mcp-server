package helper

import (
	"encoding/json"
	"net/http"
)

// DecodeJSONRequest decodes JSON request body
func DecodeJSONRequest(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// WriteJSONResponse writes JSON response with proper headers
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
