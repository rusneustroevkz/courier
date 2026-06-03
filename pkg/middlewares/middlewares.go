package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rusneustroevkz/courier/pkg/responder"
)

type Config struct {
	Env string
}

type contextKey string

const UserIDKey contextKey = "user_id"

type Middleware interface {
	Auth(next http.Handler) http.Handler
	CORS(next http.Handler) http.Handler
	RequestID(next http.Handler) http.Handler
	RestorePanics(next http.Handler) http.Handler
}

type middleware struct {
	cfg         Config
	authService Auth
}

func NewMiddleware(cfg Config, authService Auth) Middleware {
	return &middleware{
		cfg:         cfg,
		authService: authService,
	}
}

type Response struct {
	Errors map[string]string `json:"errors,omitempty"`
}

func (m *middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := uuid.NewV7()
		if err != nil {
			slog.ErrorContext(r.Context(), "failed to generate request ID", "err", err)
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "request_id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := slog.With("method", "Auth", "path", r.URL.Path)
		res := &Response{
			Errors: make(map[string]string),
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.ErrorContext(r.Context(), "no have auth header")
			res.Errors["error"] = "Нет авторизационного заголовка"
			responder.Responder(w, res, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.ErrorContext(r.Context(), "invalid token format")
			res.Errors["error"] = "Невалидный формат токена"
			responder.Responder(w, res, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		userID, err := m.authService.VerifyToken(tokenString)
		if err != nil {
			log.ErrorContext(r.Context(), "invalid or expired token")
			res.Errors["error"] = "Невалидный или истекший токен"
			responder.Responder(w, res, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if strings.Contains(m.cfg.Env, "prd") {
			w.Header().Set("Access-Control-Allow-Origin", "https://b2b-courier-14.ru")
		}
		if strings.Contains(m.cfg.Env, "local") {
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "ngrok-free.app") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Headers,Origin,Accept,X-Requested-With,Content-Type,Access-Control-Request-Method,Access-Control-Request-Headers,Authorization,X-Service,ngrok-skip-browser-warning",
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
				slog.Error("received panic", "err", err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
