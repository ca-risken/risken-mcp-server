package helper

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadAndRestoreRequestBody(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "valid json body",
			body: `{"id": "test-id", "method": "test"}`,
		},
		{
			name: "empty body",
			body: "",
		},
		{
			name: "large body",
			body: `{"data": "` + string(make([]byte, 1000)) + `"}`,
		},
		{
			name: "unicode content",
			body: `{"message": "こんにちは世界"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			got, err := ReadAndRestoreRequestBody(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadAndRestoreRequestBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check if the returned bytes match the original body
				if string(got) != tt.body {
					t.Errorf("ReadAndRestoreRequestBody() got = %v, want %v\ndiff: %s", string(got), tt.body, cmp.Diff(tt.body, string(got)))
				}

				// Check if the request body can be read again
				restoredBytes, err := io.ReadAll(req.Body)
				if err != nil {
					t.Errorf("failed to read restored body: %v", err)
				}

				if string(restoredBytes) != tt.body {
					t.Errorf("restored body mismatch got = %v, want %v\ndiff: %s", string(restoredBytes), tt.body, cmp.Diff(tt.body, string(restoredBytes)))
				}
			}
		})
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authHeader    string
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			authHeader:    "Bearer valid-token-123",
			expectedToken: "valid-token-123",
		},
		{
			name:          "empty header",
			authHeader:    "",
			expectedToken: "",
		},
		{
			name:          "invalid format",
			authHeader:    "Basic token123",
			expectedToken: "",
		},
		{
			name:          "bearer without token",
			authHeader:    "Bearer ",
			expectedToken: "",
		},
		{
			name:          "bearer with spaces",
			authHeader:    "Bearer token with spaces",
			expectedToken: "token with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			got := ExtractBearerToken(req)
			if got != tt.expectedToken {
				t.Errorf("ExtractBearerToken() = %v, want %v", got, tt.expectedToken)
			}
		})
	}
}
