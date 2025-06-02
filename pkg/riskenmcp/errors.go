package riskenmcp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

// JSONRPCResponse defines the response object for JSON-RPC 2.0.
// @see https://www.jsonrpc.org/specification#response_object
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"` // "2.0"
	ID      any           `json:"id"`      // string or number or null
	Error   *JSONRPCError `json:"error,omitempty"`
	// Result  any           `json:"result,omitempty"`
}

// JSONRPCError defines the error object for JSON-RPC 2.0.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	// Standard errors
	JSONRPCErrorParseError     = -32700
	JSONRPCErrorInvalidRequest = -32600
	JSONRPCErrorMethodNotFound = -32601
	JSONRPCErrorInvalidParams  = -32602
	JSONRPCErrorInternalError  = -32603

	// Custom errors (-32000 ~ -32099)
	JSONRPCErrorUnauthorized = -32001
)

func NewJSONRPCError(id any, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      &id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}

func (j *JSONRPCResponse) String() string {
	jsonBytes, _ := json.Marshal(j)
	return string(jsonBytes)
}

type JSONRPCRequest struct {
	ID any `json:"id"`
}

func ParseJSONRPCRequestID(r *http.Request) (any, error) {
	bodyBytes, err := helper.ReadAndRestoreRequestBody(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	if len(bodyBytes) == 0 {
		return nil, nil
	}

	jsonRPC := JSONRPCRequest{}
	err = json.Unmarshal(bodyBytes, &jsonRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
	}
	// Convert numeric types to int according to JSON-RPC 2.0 spec
	switch id := jsonRPC.ID.(type) {
	case float64:
		return int(id), nil
	case int:
		return id, nil
	case string:
		if id == "" {
			return nil, nil
		}
		return id, nil
	case nil:
		return nil, nil
	default:
		return jsonRPC.ID, nil
	}
}
