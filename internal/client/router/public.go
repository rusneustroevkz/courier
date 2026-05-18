package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rusneustroevkz/courier/internal/client/auth"
	"github.com/rusneustroevkz/courier/internal/client/users"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
)

type Public interface {
	Routes() *chi.Mux
}

type public struct {
	mw              middlewares.Middleware
	usersController users.Controller
	authController  auth.Controller
}

func NewPublic(mw middlewares.Middleware, usersController users.Controller, authController auth.Controller) Public {
	return &public{
		mw:              mw,
		usersController: usersController,
		authController:  authController,
	}
}

func (rr *public) Routes() *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(rr.mw.CORS)
	mux.Use(rr.mw.RequestID)
	mux.Use(rr.mw.RestorePanics)

	mux.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", rr.authController.Register)
			r.Post("/login", rr.authController.Login)
			r.With(rr.mw.Auth).Post("/refresh", rr.authController.Refresh)
			r.With(rr.mw.Auth).Post("/logout", rr.authController.Logout)
		})
		r.With(rr.mw.Auth).Route("/users", func(r chi.Router) {
			r.Get("/me", rr.usersController.GetMe)
		})
	})

	return mux
}
