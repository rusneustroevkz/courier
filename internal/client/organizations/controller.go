package organizations

import (
	"database/sql"
	"encoding/json"
	"github.com/rusneustroevkz/courier/internal/client/orders"
	"github.com/rusneustroevkz/courier/internal/client/users"
	"github.com/rusneustroevkz/courier/pkg/responder"
	"github.com/rusneustroevkz/courier/pkg/utils"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Controller interface {
	My(w http.ResponseWriter, r *http.Request)
	Transactions(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	organizationsService Service
	ordersService        orders.Service
	usersService         users.Service
}

func NewController(organizationsService Service, ordersService orders.Service, usersService users.Service) Controller {
	return &controller{
		organizationsService: organizationsService,
		ordersService:        ordersService,
		usersService:         usersService,
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

type TransactionsRequest struct {
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
}
type TransactionsResponse struct {
	Errors     map[string]string       `json:"errors,omitempty"`
	Data       []TransactionsData      `json:"data,omitempty"`
	Pagination *TransactionsPagination `json:"pagination,omitempty"`
}
type TransactionsPagination struct {
	Page       int64 `json:"page,omitempty"`        // Текущая страница
	PageSize   int64 `json:"page_size,omitempty"`   // Размер страницы
	TotalCount int64 `json:"total_count,omitempty"` // Всего записей в БД по этому фильтру
	TotalPages int64 `json:"total_pages,omitempty"`
}
type TransactionsData struct {
	ID          int64     `json:"id"`
	AddressFrom string    `json:"address_from"`
	AddressTo   string    `json:"address_to"`
	CreatedAt   time.Time `json:"created_at"`
	Price       float64   `json:"price"`
}

// Transactions Список транзакций баланса
//
//	@Summary      Список транзакций баланса
//	@Description  Список транзакций баланса
//	@Tags         organizations
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body TransactionsRequest true "тело запроса"
//	@Success      200  {object} TransactionsResponse
//	@Failure      400  {object} TransactionsResponse
//	@Failure      401  {object} TransactionsResponse
//	@Failure      404  {object} TransactionsResponse
//	@Failure      500  {object} TransactionsResponse
//	@Router       /organizations/transactions [post]
func (c *controller) Transactions(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsTransactions")

	res := &TransactionsResponse{
		Errors: make(map[string]string),
	}

	var req TransactionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(r.Context(), "failed decode request", "error", err)
		res.Errors["message"] = "Ошибка обработки запроса"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	userID, err := utils.GetFromCtx(r.Context())
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user id from context", "error", err)
		res.Errors["message"] = "В контексте отсутствует идентификатор пользователя"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}
	user, err := c.usersService.GetByID(r.Context(), userID)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user", "error", err)
		res.Errors["message"] = "Ошибка выборки пользователя"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	if !user.OrganizationID.Valid {
		res.Errors["message"] = "Пользователь не привязан к организации"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	params := orders.GetAll{
		OrganizationID: user.OrganizationID,
		Status:         sql.NullString{String: string(orders.OrderStatusDelivered), Valid: true},
		Page:           req.Page,
		PageSize:       req.PageSize,
	}
	result, err := c.ordersService.GetAll(r.Context(), params)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get orders", "error", err)
		res.Errors["message"] = "Ошибка списка заказов"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Pagination = &TransactionsPagination{
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	}
	for _, order := range result.Data {
		item := TransactionsData{
			ID:          order.ID,
			AddressFrom: order.FromAddress,
			AddressTo:   order.ToAddress,
			CreatedAt:   order.CreatedAt,
		}
		if order.Price.Valid {
			price, err := strconv.ParseFloat(order.Price.String, 64)
			if err != nil {
				log.ErrorContext(r.Context(), "failed parse price", "error", err)
			} else {
				item.Price = price
			}
		}

		res.Data = append(res.Data, item)
	}

	responder.Responder(w, res, http.StatusOK)
}
