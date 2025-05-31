package main

import (
	"fmt"
	"os"

	"github.com/ca-risken/go-risken"
)

func newRISKENClient() (*risken.Client, error) {
	token := os.Getenv("RISKEN_ACCESS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("RISKEN_ACCESS_TOKEN not set")
	}
	url := os.Getenv("RISKEN_URL")
	if url == "" {
		return nil, fmt.Errorf("RISKEN_URL not set")
	}
	riskenClient := risken.NewClient(token, risken.WithAPIEndpoint(url))
	return riskenClient, nil
}
