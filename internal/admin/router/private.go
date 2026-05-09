package router

import "github.com/go-chi/chi/v5"

type Private interface {
	Routes() *chi.Mux
}

type private struct{}

func NewPrivate() Private {
	return &private{}
}

func (r *private) Routes() *chi.Mux {
	mux := chi.NewRouter()

	return mux
}
