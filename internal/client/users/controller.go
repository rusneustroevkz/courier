package users

import (
	"net/http"
	"strconv"

	"github.com/rusneustroevkz/courier/pkg/logger"
	"github.com/rusneustroevkz/courier/pkg/responder"
	"github.com/rusneustroevkz/courier/pkg/utils"
)

type Controller interface {
	GetMe(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	usersService Service
}

func NewController(usersService Service) Controller {
	return &controller{
		usersService: usersService,
	}
}

type GetMeResponse struct {
	Errors map[string]string  `json:"errors,omitempty"`
	Data   *GetMeResponseData `json:"data,omitempty"`
}
type GetMeResponseData struct {
	ID       int64   `json:"id,omitempty"`
	TgID     int64   `json:"tg_id,omitempty"`
	FullName string  `json:"full_name,omitempty"`
	Email    string  `json:"email,omitempty"`
	Phone    string  `json:"phone,omitempty"`
	Role     string  `json:"role,omitempty"`
	OnWork   bool    `json:"on_work,omitempty"`
	Verified bool    `json:"verified,omitempty"`
	Rating   float64 `json:"rating,omitempty"`
	Balance  float64 `json:"balance,omitempty"`
}

// GetMe Данные о профиле
//
//	@Summary      Данные о профиле
//	@Description  Данные о профиле
//	@Tags         users
//	@Accept       application/json
//	@Produce      application/json
//	@Success      200  {object} GetMeResponse
//	@Failure      400  {object} GetMeResponse
//	@Failure      401  {object} GetMeResponse
//	@Failure      404  {object} GetMeResponse
//	@Failure      500  {object} GetMeResponse
//	@Router       /users/me [get]
func (c *controller) GetMe(w http.ResponseWriter, r *http.Request) {
	log := logger.With("method", "Login")

	res := &GetMeResponse{
		Errors: make(map[string]string),
	}

	userID, err := utils.GetFromCtx(r.Context())
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user id from context", "error", err)
		res.Errors["error"] = "В контексте отсутствует идентификатор пользователя"
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}
	user, err := c.usersService.GetByID(r.Context(), userID)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user", "error", err)
		res.Errors["error"] = "Ошибка выборки пользователя"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &GetMeResponseData{
		ID:       user.ID,
		Role:     string(user.Role),
		OnWork:   user.OnWork,
		Verified: user.Verified,
	}
	if user.TgID.Valid {
		res.Data.TgID = user.TgID.Int64
	}
	if user.FullName.Valid {
		res.Data.FullName = user.FullName.String
	}
	if user.Email.Valid {
		res.Data.Email = user.Email.String
	}
	if user.Phone.Valid {
		res.Data.Phone = user.Phone.String
	}
	if user.Rating.Valid {
		rating, err := strconv.ParseFloat(user.Rating.String, 10)
		if err != nil {
			log.ErrorContext(r.Context(), "failed parse rating", "error", err)
			res.Errors["rating"] = err.Error()
		} else {
			res.Data.Rating = rating
		}
	}
	if user.Balance.Valid {
		balance, err := strconv.ParseFloat(user.Balance.String, 10)
		if err != nil {
			log.ErrorContext(r.Context(), "failed parse balance", "error", err)
			res.Errors["balance"] = err.Error()
		} else {
			res.Data.Balance = balance
		}
	}

	responder.Responder(w, res, http.StatusOK)
}
