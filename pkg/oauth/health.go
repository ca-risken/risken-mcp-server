package oauth

import (
	"log/slog"
	"net/http"
)

func (s *Server) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		s.logger.Error("Failed to write healthz response", slog.String("error", err.Error()))
	}
}
