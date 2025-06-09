package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HTTPClient represents a helper for making HTTP requests
type HTTPClient struct {
	client *http.Client
	logger *slog.Logger
}

// NewHTTPClient creates a new HTTP client helper
func NewHTTPClient(logger *slog.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// JSONRequest represents a JSON HTTP request
type JSONRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    any
	LogName string // For logging purposes
}

// JSONResponse represents a JSON HTTP response
type JSONResponse struct {
	StatusCode int
	Body       map[string]any
	RawBody    string
	Headers    http.Header
}

// DoJSONRequest performs a JSON HTTP request with comprehensive logging
func (h *HTTPClient) DoJSONRequest(ctx context.Context, req JSONRequest) (*JSONResponse, error) {
	// Handle different body types
	var requestBody []byte
	var err error

	switch body := req.Body.(type) {
	case string:
		requestBody = []byte(body)
	case nil:
		requestBody = nil
	default:
		requestBody, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Log request details
	h.logger.Debug("HTTP request",
		slog.String("operation", req.LogName),
		slog.String("method", req.Method),
		slog.String("url", req.URL))

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set default Content-Type for JSON if not specified and body exists
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		if _, isString := req.Body.(string); !isString {
			httpReq.Header.Set("Content-Type", "application/json")
		}
	}

	// Perform request
	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response if successful
	var jsonBody map[string]any
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &jsonBody); err != nil {
			return nil, fmt.Errorf("failed to decode JSON response: %w", err)
		}
	}

	return &JSONResponse{
		StatusCode: resp.StatusCode,
		Body:       jsonBody,
		RawBody:    string(responseBody),
		Headers:    resp.Header,
	}, nil
}

// DoJSONRequestWithValidation performs JSON request and validates success status
func (h *HTTPClient) DoJSONRequestWithValidation(ctx context.Context, req JSONRequest) (*JSONResponse, error) {
	resp, err := h.DoJSONRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.RawBody)
	}

	return resp, nil
}

// DoSimpleGET performs a simple GET request for non-JSON responses
func (h *HTTPClient) DoSimpleGET(ctx context.Context, url, logName string) ([]byte, error) {
	h.logger.Debug("HTTP GET request",
		slog.String("operation", logName),
		slog.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	h.logger.Debug("HTTP GET response",
		slog.String("operation", logName),
		slog.Int("status_code", resp.StatusCode),
		slog.Int("body_size", len(body)))

	return body, nil
}
