package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rusneustroevkz/courier/internal/client/middlewares"
	"github.com/rusneustroevkz/courier/internal/client/users"
)

type Public interface {
	Routes() *chi.Mux
}

type public struct {
	usersController users.Controller
	mw              middlewares.Middleware
}

func NewPublic(mw middlewares.Middleware, usersController users.Controller) Public {
	return &public{
		usersController: usersController,
		mw:              mw,
	}
}

func (rr *public) Routes() *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(rr.mw.CORS)
	mux.Use(rr.mw.RequestID)
	mux.Use(rr.mw.RestorePanics)

	mux.Route("/api/v1", func(r chi.Router) {
		r.With(rr.mw.Auth).Route("/users", func(r chi.Router) {
			r.Get("/", rr.usersController.GetByID)
		})
	})

	return mux
}
