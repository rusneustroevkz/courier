package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rusneustroevkz/courier/internal/config"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type server struct {
	srv *http.Server
}

func New(cfg config.Server, router chi.Router) Server {
	return &server{
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: router,
		},
	}
}

func (s *server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
