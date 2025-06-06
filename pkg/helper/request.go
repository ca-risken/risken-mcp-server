package helper

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func ReadAndRestoreRequestBody(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyBytes, nil
}

func ExtractBearerToken(r *http.Request) string {
	token := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	return token
}

func ExtractRISKENTokenFromHeader(r *http.Request) string {
	token := ""
	authHeader := r.Header.Get("RISKEN-ACCESS-TOKEN")
	if authHeader != "" {
		token = authHeader
	}
	return token
}

// ExtractClientIP extracts client IP from request
func ExtractClientIP(r *http.Request) string {
	// X-Forwarded-For header check
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx]) // Take the first IP
		}
		return strings.TrimSpace(xff)
	}

	// X-Real-IP header check
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Remote address fallback
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}
