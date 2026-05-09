package router

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	client "github.com/rusneustroevkz/courier/frontend/client"
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

	mux.Handle("/", templ.Handler(client.Index()))
	mux.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	return mux
}
