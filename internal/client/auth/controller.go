package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"github.com/rusneustroevkz/courier/pkg/responder"
)

const (
	refreshTokenName = "refresh_token"
)

type Controller interface {
	Register(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	authService Service
	validate    *validator.Validate
}

func NewController(authService Service) Controller {
	validate := validator.New()

	c := &controller{
		authService: authService,
		validate:    validate,
	}

	c.registerTagNameFunc(validate)

	return c
}

type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,gte=8,lte=30"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
}
type RegisterResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
}

// Register Регистрация пользователя
//
//	@Summary      Регистрация пользователя
//	@Description  Регистрация пользователя
//	@Tags         auth
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body RegisterRequest true "тело запроса"
//	@Success      201  {object} RegisterResponse
//	@Failure      400  {object} RegisterResponse
//	@Failure      404  {object} RegisterResponse
//	@Failure      500  {object} RegisterResponse
//	@Router       /auth/register [post]
func (c *controller) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	log := logger.With("method", "Register")
	res := &RegisterResponse{
		Errors: make(map[string]string),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(r.Context(), "failed decode request", "error", err)
		res.Errors["error"] = "Ошибка обработки запроса"
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	if errorsMap, err := c.Struct(req); errorsMap != nil {
		log.ErrorContext(r.Context(), "failed validate request", "err", err)
		res.Errors["error"] = "Ошибка валидации параметров"
		for key, val := range errorsMap {
			res.Errors[key] = val
		}
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	params := Register{
		Email:    req.Email,
		Password: req.Password,
	}
	if err := c.authService.Register(r.Context(), params); err != nil {
		log.ErrorContext(r.Context(), "failed register", "error", err)
		res.Errors["error"] = "Ошибка регистрации пользователя: " + err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	responder.Responder(w, res, http.StatusOK)
}

type RefreshResponse struct {
	Errors map[string]string    `json:"errors,omitempty"`
	Data   *RefreshResponseData `json:"data,omitempty"`
}
type RefreshResponseData struct {
	AccessToken string `json:"access_token"`
}

// Refresh Обновление токена
//
//	@Summary      Обновление токена
//	@Description  Обновление токена
//	@Tags         auth
//	@Accept       application/json
//	@Produce      application/json
//	@Success      201  {object} RefreshResponse
//	@Failure      400  {object} RefreshResponse
//	@Failure      404  {object} RefreshResponse
//	@Failure      500  {object} RefreshResponse
//	@Router       /auth/refresh [post]
func (c *controller) Refresh(w http.ResponseWriter, r *http.Request) {
	log := logger.With("method", "Refresh")
	res := &RefreshResponse{
		Errors: make(map[string]string),
	}

	cookie, err := r.Cookie(refreshTokenName)
	if err != nil {
		log.ErrorContext(r.Context(), "failed decode request", "error", err)
		res.Errors["error"] = "Нет авторизационной куки"
		responder.Responder(w, res, http.StatusUnauthorized)
		return
	}

	refreshToken, accessToken, err := c.authService.Refresh(r.Context(), cookie.Value)
	if err != nil {
		log.ErrorContext(r.Context(), "failed refresh token", "error", err)
		res.Errors["error"] = "Ошибка обновления токена"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &RefreshResponseData{
		AccessToken: accessToken,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenName,
		Value:    refreshToken,
		Expires:  time.Now().Add(refreshTokenTTL),
		HttpOnly: true,
		Secure:   true,
		Path:     "/api/v1/auth/refresh",
	})

	responder.Responder(w, res, http.StatusOK)
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=8,lte=30"`
}
type LoginResponse struct {
	Errors map[string]string  `json:"errors,omitempty"`
	Data   *LoginResponseData `json:"data,omitempty"`
}
type LoginResponseData struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id,omitempty"`
	UserEmail   string `json:"user_email,omitempty"`
}

// Login Вход в аккаунт
//
//	@Summary      Вход в аккаунт
//	@Description  Вход в аккаунт
//	@Tags         auth
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body LoginRequest true "тело запроса"
//	@Success      201  {object} LoginResponse
//	@Failure      400  {object} LoginResponse
//	@Failure      404  {object} LoginResponse
//	@Failure      500  {object} LoginResponse
//	@Router       /auth/login [post]
func (c *controller) Login(w http.ResponseWriter, r *http.Request) {
	log := logger.With("method", "Login")

	var req LoginRequest
	res := &LoginResponse{
		Errors: make(map[string]string),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(r.Context(), "failed decode request", "error", err)
		res.Errors["error"] = "Ошибка обработки запроса: " + err.Error()
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	if errorsMap, err := c.Struct(req); errorsMap != nil {
		log.ErrorContext(r.Context(), "failed validate request", "err", err)
		res.Errors["error"] = "Ошибка валидации параметров"
		for key, val := range errorsMap {
			res.Errors[key] = val
		}
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	user, refreshToken, accessToken, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.ErrorContext(r.Context(), "failed login", "error", err)
		res.Errors["error"] = "Ошибка авторизации"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &LoginResponseData{
		AccessToken: accessToken,
		UserID:      user.ID,
	}
	if user.Email.Valid {
		res.Data.UserEmail = user.Email.String
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenName,
		Value:    refreshToken,
		Expires:  time.Now().Add(refreshTokenTTL),
		HttpOnly: true,
		Secure:   true,
		Path:     "/api/v1/auth/refresh",
	})

	responder.Responder(w, res, http.StatusOK)
}

type LogoutResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
}

// Logout Выход из аккаунта
//
//	@Summary      Выход из аккаунта
//	@Description  Выход из аккаунта
//	@Tags         auth
//	@Accept       application/json
//	@Produce      application/json
//	@Success      201  {object} LogoutResponse
//	@Failure      400  {object} LogoutResponse
//	@Failure      404  {object} LogoutResponse
//	@Failure      500  {object} LogoutResponse
//	@Router       /auth/logout [post]
func (c *controller) Logout(w http.ResponseWriter, r *http.Request) {
	log := logger.With("method", "Logout")

	res := &LogoutResponse{
		Errors: make(map[string]string),
	}

	cookie, err := r.Cookie(refreshTokenName)
	if err != nil {
		log.ErrorContext(r.Context(), "failed logout", "error", err)
		res.Errors["error"] = "Нет авторизационной куки"
		responder.Responder(w, res, http.StatusNoContent)
		return
	}

	if err := c.authService.Logout(r.Context(), cookie.Value); err != nil {
		log.ErrorContext(r.Context(), "failed logout", "error", err)
		res.Errors["error"] = "Ошибка логаута"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenName,
		Value:    "",
		Path:     "/auth/refresh",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}
