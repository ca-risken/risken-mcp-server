package streamablehttp

import (
	"log/slog"
	"net/http"
)

func (a *AuthServer) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		a.logger.Error("Failed to write healthz response", slog.String("error", err.Error()))
	}
}
