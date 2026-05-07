package router

import (
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	web "github.com/rusneustroevkz/courier/public"
)

type Render interface {
	Routes() *chi.Mux
}

type render struct {
}

func NewRender() Render {
	return &render{}
}

func (rr *render) Routes() *chi.Mux {
	mux := chi.NewRouter()

	component := web.Hello("John")

	mux.Route("/", func(r chi.Router) {
		r.Handle("/", templ.Handler(component))
	})

	return mux
}
