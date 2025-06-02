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
