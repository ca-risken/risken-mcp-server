package main

import (
	"fmt"

	"github.com/ca-risken/go-risken"
)

func newRISKENClient(url, token string) (*risken.Client, error) {
	if url == "" {
		return nil, fmt.Errorf("RISKEN_URL not set")
	}
	if token == "" {
		return nil, fmt.Errorf("RISKEN_ACCESS_TOKEN not set")
	}
	riskenClient := risken.NewClient(token, risken.WithAPIEndpoint(url))
	return riskenClient, nil
}
