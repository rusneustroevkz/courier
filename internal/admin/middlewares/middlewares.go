package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rusneustroevkz/courier/internal/admin/config"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"github.com/rusneustroevkz/courier/pkg/responder"
)

type Middleware interface {
	Auth(next http.Handler) http.Handler
	CORS(next http.Handler) http.Handler
	RequestID(next http.Handler) http.Handler
	RestorePanics(next http.Handler) http.Handler
}

type middleware struct {
	cfg *config.Config
}

func NewMiddleware(cfg *config.Config) Middleware {
	return &middleware{
		cfg: cfg,
	}
}

type Response struct {
	Errors map[string]string `json:"errors,omitempty"`
}

func (m *middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := uuid.NewV7()
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to generate request ID", "err", err)
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "request_id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(m.cfg.ENV, "prd") {
			w.Header().Set("Access-Control-Allow-Origin", "https://b2b-courier-14.ru")
		}
		if strings.Contains(m.cfg.ENV, "local") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Headers,Origin,Accept,X-Requested-With,Content-Type,Access-Control-Request-Method,Access-Control-Request-Headers,Authorization,X-Service",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *middleware) RestorePanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := Response{
			Errors: make(map[string]string),
		}

		defer func() {
			if err := recover(); err != nil {
				responder.Responder(w, nil, http.StatusInternalServerError)
				res.Errors["error"] = fmt.Sprintf("Паника: %v", err)
				logger.Error("received panic", "err", err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
