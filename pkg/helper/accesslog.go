package helper

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// UseAccessLogging creates an access logging middleware
func UseAccessLogging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			bodyBytes, _ := ReadAndRestoreRequestBody(r)

			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
			next.ServeHTTP(wrapped, r)

			// Log response
			duration := time.Since(start)
			AccessLogging(r, logger, wrapped.statusCode, duration, bodyBytes)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.headerWritten {
		return
	}
	rw.statusCode = code
	rw.headerWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(200)
	}
	return rw.ResponseWriter.Write(data)
}

func AccessLogging(r *http.Request, logger *slog.Logger, statusCode int, duration time.Duration, bodyBytes []byte) {
	logger.Info("AccessLog",
		slog.String("type", "access_log"),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("content_type", r.Header.Get("Content-Type")),
		slog.String("mcp_session_id", r.Header.Get("Mcp-Session-Id")),
		slog.Int("status", statusCode),
		slog.String("duration", fmt.Sprintf("%.3fms", float64(duration.Nanoseconds())/1e6)),
		slog.String("client_ip", ExtractClientIP(r)),
		slog.String("user_agent", r.UserAgent()),
	)

	// for debug
	body := string(bodyBytes)
	logger.Debug("RequestDetail",
		slog.String("type", "request_body"),
		slog.String("json_rpc", body),
		slog.String("mcp_session_id", r.Header.Get("Mcp-Session-Id")),
		slog.Int("status", statusCode),
	)
}
