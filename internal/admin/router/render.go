package router

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	admin "github.com/rusneustroevkz/courier/static/admin"
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

	mux.Handle("/", templ.Handler(admin.Index()))
	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return mux
}
