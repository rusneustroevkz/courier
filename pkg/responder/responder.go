package responder

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Responder(w http.ResponseWriter, data any, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to marshal response", "err", err)
	}
}

func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}
