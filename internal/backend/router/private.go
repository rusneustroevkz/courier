package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Private interface {
	Routes() *chi.Mux
}

type private struct{}

func NewPrivate() Private {
	return &private{}
}

func (r *private) Routes() *chi.Mux {
	mux := chi.NewRouter()

	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return mux
}
