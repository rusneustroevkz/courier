package router

import "github.com/go-chi/chi/v5"

type Public interface {
	Routes() *chi.Mux
}

type public struct{}

func NewPublic() Public {
	return &public{}
}

func (r *public) Routes() *chi.Mux {
	mux := chi.NewRouter()

	return mux
}
