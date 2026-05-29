package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rusneustroevkz/courier/internal/client/auth"
	"github.com/rusneustroevkz/courier/internal/client/orders"
	"github.com/rusneustroevkz/courier/internal/client/organizations_branches"
	"github.com/rusneustroevkz/courier/internal/client/users"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
)

type Public interface {
	Routes() *chi.Mux
}

type public struct {
	mw                              middlewares.Middleware
	usersController                 users.Controller
	authController                  auth.Controller
	ordersController                orders.Controller
	organizationsBranchesController organizations_branches.Controller
}

func NewPublic(
	mw middlewares.Middleware,
	usersController users.Controller,
	authController auth.Controller,
	ordersController orders.Controller,
	organizationsBranchesController organizations_branches.Controller,
) Public {
	return &public{
		mw:                              mw,
		usersController:                 usersController,
		authController:                  authController,
		ordersController:                ordersController,
		organizationsBranchesController: organizationsBranchesController,
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
		r.With(rr.mw.Auth).Route("/orders", func(r chi.Router) {
			r.Post("/", rr.ordersController.Create)
			r.Post("/list", rr.ordersController.GetAll)
			r.Get("/{id}", rr.ordersController.GetByID)
			r.Put("/", rr.ordersController.UpdateCourier)
			r.Post("/courier-location", rr.ordersController.GetCourierLocation)
		})
		r.With(rr.mw.Auth).Route("/organizations-branches", func(r chi.Router) {
			r.Post("/", rr.organizationsBranchesController.Create)
			r.Post("/list", rr.organizationsBranchesController.List)
			r.Get("/{id}", rr.organizationsBranchesController.GetByID)
			r.Post("/suggest", rr.organizationsBranchesController.Suggest)
			r.Post("/set-activation", rr.organizationsBranchesController.SetActivation)
			r.Post("/set-user-selected", rr.organizationsBranchesController.SetUserSelected)
		})
	})

	return mux
}
