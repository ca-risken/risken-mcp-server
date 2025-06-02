package riskenmcp

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseJSONRPCRequestID(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    any
		wantErr bool
	}{
		{
			name: "string id",
			body: `{"id": "test-id", "method": "test"}`,
			want: "test-id",
		},
		{
			name: "numeric id",
			body: `{"id": 123, "method": "test"}`,
			want: 123,
		},
		{
			name: "empty id",
			body: `{"id": "", "method": "test"}`,
			want: nil,
		},
		{
			name: "zero id",
			body: `{"id": 0, "method": "test"}`,
			want: 0,
		},
		{
			name: "json without id field",
			body: `{"method": "test"}`,
			want: nil,
		},
		{
			name: "empty body",
			body: "",
			want: nil,
		},
		{
			name:    "invalid json",
			body:    `{"id": "test-id", "method":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			got, err := ParseJSONRPCRequestID(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONRPCRequestID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseJSONRPCRequestID() got = %v, want %v\ndiff: %s", got, tt.want, cmp.Diff(tt.want, got))
			}
		})
	}
}
