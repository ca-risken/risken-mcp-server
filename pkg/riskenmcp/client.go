package riskenmcp

import (
	"context"
	"fmt"

	"github.com/ca-risken/go-risken"
)

type contextKey string

const (
	RISKENClientContextKey contextKey = "risken_client"
)

// WithRISKENClient sets the RISKEN client in the context.
func WithRISKENClient(ctx context.Context, client *risken.Client) context.Context {
	return context.WithValue(ctx, RISKENClientContextKey, client)
}

// GetRISKENClient returns the RISKEN client from server field or context.
func (s *Server) GetRISKENClient(ctx context.Context) (*risken.Client, error) {
	if s.riskenClient != nil {
		return s.riskenClient, nil
	}

	client, ok := ctx.Value(RISKENClientContextKey).(*risken.Client)
	if !ok || client == nil {
		return nil, fmt.Errorf("no RISKEN client found in context")
	}
	return client, nil
}
