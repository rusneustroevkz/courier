package organizations

import (
	"github.com/rusneustroevkz/courier/pkg/responder"
	"github.com/rusneustroevkz/courier/pkg/utils"
	"log/slog"
	"net/http"
	"strconv"
)

type Controller interface {
	My(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	organizationsService Service
}

func NewController(organizationsService Service) Controller {
	return &controller{
		organizationsService: organizationsService,
	}
}

type MyResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
	Data   *MyData           `json:"data,omitempty"`
}
type MyData struct {
	ID      int64   `json:"id,omitempty"`
	Name    string  `json:"name,omitempty"`
	Balance float64 `json:"balance,omitempty"`
}

// My Выборка своей организации
//
//	@Summary      Выборка своей организации
//	@Description  Выборка своей организации
//	@Tags         organizations
//	@Accept       application/json
//	@Produce      application/json
//	@Success      200  {object} MyResponse
//	@Failure      400  {object} MyResponse
//	@Failure      401  {object} MyResponse
//	@Failure      404  {object} MyResponse
//	@Failure      500  {object} MyResponse
//	@Router       /organizations/my [get]
func (c *controller) My(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsMy")

	res := MyResponse{
		Errors: make(map[string]string),
	}

	userID, err := utils.GetFromCtx(r.Context())
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user id from context", "error", err)
		res.Errors["message"] = "В контексте отсутствует идентификатор пользователя"
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	organization, err := c.organizationsService.My(r.Context(), userID)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get organization", "error", err)
		res.Errors["message"] = "Ошибка выборки организации"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &MyData{
		ID:   organization.ID,
		Name: organization.Name,
	}
	balance, err := strconv.ParseFloat(organization.Balance, 64)
	if err != nil {
		log.ErrorContext(r.Context(), "failed parse balance", "error", err)
		res.Errors["message"] = "Ошибка расчета баланса"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}
	res.Data.Balance = balance

	responder.Responder(w, res, http.StatusOK)
}
