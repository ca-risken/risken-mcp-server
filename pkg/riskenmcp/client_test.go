package riskenmcp

import (
	"context"
	"testing"

	"github.com/ca-risken/go-risken"
)

func TestGetRISKENClient(t *testing.T) {
	tests := []struct {
		name    string
		client  *risken.Client
		ctx     context.Context
		want    *risken.Client
		wantErr bool
	}{
		{
			name:   "server has client",
			client: &risken.Client{},
			ctx:    context.Background(),
			want:   &risken.Client{},
		},
		{
			name:   "context has client",
			client: nil,
			ctx:    WithRISKENClient(context.Background(), &risken.Client{}),
			want:   &risken.Client{},
		},
		{
			name:    "no client",
			client:  nil,
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				riskenClient: tt.client,
			}

			got, err := s.GetRISKENClient(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRISKENClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("GetRISKENClient() got = nil, want non-nil client")
			}
		})
	}
}
