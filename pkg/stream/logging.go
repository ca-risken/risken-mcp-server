package stream

import (
	"log/slog"
	"net/http"

	"github.com/ca-risken/risken-mcp-server/pkg/helper"
)

func (a *AuthStreamableHTTPServer) accessLogging(r *http.Request) {
	jsonRPC := ""
	bodyBytes, err := helper.ReadAndRestoreRequestBody(r)
	if err == nil {
		jsonRPC = string(bodyBytes)
	}

	a.logger.Debug("Received request",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("content_type", r.Header.Get("Content-Type")),
		slog.String("mcp_session_id", r.Header.Get("Mcp-Session-Id")),
		slog.String("json_rpc", jsonRPC),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
	)
}
