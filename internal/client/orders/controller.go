package orders

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rusneustroevkz/courier/internal/client/users"
	"github.com/rusneustroevkz/courier/pkg/responder"
	"github.com/rusneustroevkz/courier/pkg/utils"
)

type Controller interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	UpdateCourier(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	usersService  users.Service
	ordersService Service
}

func NewController(usersService users.Service, ordersService Service) Controller {
	return &controller{
		usersService:  usersService,
		ordersService: ordersService,
	}
}

type CreateRequest struct {
	FromAddress string  `json:"from_address" validate:"required"`
	FromLat     float64 `json:"from_lat" validate:"required"`
	FromLon     float64 `json:"from_lon" validate:"required"`
	ToAddress   string  `json:"to_address" validate:"required"`
	ToLat       float64 `json:"to_lat" validate:"required"`
	ToLon       float64 `json:"to_lon" validate:"required"`
	Price       float64 `json:"price" validate:"required"`
	Description string  `json:"description" validate:"omitempty"`
}
type CreateResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
	Data   *CreateData       `json:"data,omitempty"`
}
type CreateData struct {
	OrderID int64 `json:"order_id,omitempty"`
}

// Create Создание заказа
//
//	@Summary      Создание заказа
//	@Description  Создание заказа
//	@Tags         orders
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body CreateRequest true "тело запроса"
//	@Success      200  {object} CreateResponse
//	@Failure      400  {object} CreateResponse
//	@Failure      401  {object} CreateResponse
//	@Failure      404  {object} CreateResponse
//	@Failure      500  {object} CreateResponse
//	@Router       /orders [post]
func (c *controller) Create(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrderCreate")

	res := &CreateResponse{
		Errors: make(map[string]string),
	}

	var req CreateRequest
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
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}
	user, err := c.usersService.GetByID(r.Context(), userID)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user", "error", err)
		res.Errors["message"] = "Ошибка выборки пользователя"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	if !user.OrganizationID.Valid {
		res.Errors["message"] = "Пользователь не привязан к организации"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	params := Create{
		OrganizationID: user.OrganizationID.Int64,
		FromAddress:    req.FromAddress,
		FromLat:        req.FromLat,
		FromLon:        req.FromLon,
		ToAddress:      req.ToAddress,
		ToLat:          req.ToLat,
		ToLon:          req.ToLon,
		Price:          req.Price,
		Description:    req.Description,
	}
	id, err := c.ordersService.Create(r.Context(), params)
	if err != nil {
		log.ErrorContext(r.Context(), "failed create orders", "error", err)
		res.Errors["message"] = "Не удалось создать заказ"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &CreateData{
		OrderID: id,
	}

	responder.Responder(w, res, http.StatusOK)
}

type GetByIDResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
	Data   *GetByIDData      `json:"data,omitempty"`
}
type GetByIDData struct {
	ID              int64     `json:"id"`
	Description     string    `json:"description"`
	OrganizationID  int64     `json:"organization_id"`
	CourierID       int64     `json:"courier_id"`
	Status          string    `json:"status"`
	FromAddress     string    `json:"from_address"`
	FromLat         string    `json:"from_lat"`
	FromLon         string    `json:"from_lon"`
	ToAddress       string    `json:"to_address"`
	ToLat           string    `json:"to_lat"`
	ToLon           string    `json:"to_lon"`
	Price           string    `json:"price"`
	TgClientChatID  int64     `json:"tg_client_chat_id"`
	TgLiveMessageID int64     `json:"tg_live_message_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GetByID Выборка заказа
//
//	@Summary      Выборка заказа
//	@Description  Выборка заказа
//	@Tags         orders
//	@Accept       application/json
//	@Produce      application/json
//	@Param        id path int true "Идентификатор"
//	@Success      200  {object} GetByIDResponse
//	@Failure      400  {object} GetByIDResponse
//	@Failure      401  {object} GetByIDResponse
//	@Failure      404  {object} GetByIDResponse
//	@Failure      500  {object} GetByIDResponse
//	@Router       /orders/{id} [get]
func (c *controller) GetByID(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "GetByID")

	res := &GetByIDResponse{
		Errors: make(map[string]string),
	}

	idString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.ErrorContext(r.Context(), "invalid id", "err", err, "id", idString)
		res.Errors["message"] = "Невалидный идентификатор"
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	userID, err := utils.GetFromCtx(r.Context())
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user id from context", "error", err)
		res.Errors["message"] = "В контексте отсутствует идентификатор пользователя"
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}
	user, err := c.usersService.GetByID(r.Context(), userID)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get user", "error", err)
		res.Errors["message"] = "Ошибка выборки пользователя"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	if !user.OrganizationID.Valid {
		res.Errors["message"] = "Пользователь не привязан к организации"
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	params := GetByID{
		ID:             id,
		OrganizationID: user.OrganizationID.Int64,
	}
	order, err := c.ordersService.GetByID(r.Context(), params)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get order", "error", err)
		res.Errors["message"] = "Ошибка выборки заказа"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &GetByIDData{
		ID:             order.ID,
		OrganizationID: order.OrganizationID,
		Status:         string(order.Status),
		FromAddress:    order.FromAddress,
		FromLat:        order.FromLat,
		FromLon:        order.FromLon,
		ToAddress:      order.ToAddress,
		ToLat:          order.ToLat,
		ToLon:          order.ToLon,
		Price:          order.Price,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}

	if order.Description.Valid {
		res.Data.Description = order.Description.String
	}
	if order.CourierID.Valid {
		res.Data.CourierID = order.CourierID.Int64
	}
	if order.TgClientChatID.Valid {
		res.Data.TgClientChatID = order.TgClientChatID.Int64
	}
	if order.TgLiveMessageID.Valid {
		res.Data.TgLiveMessageID = order.TgLiveMessageID.Int64
	}

	responder.Responder(w, res, http.StatusOK)
}

func (c *controller) GetAll(w http.ResponseWriter, r *http.Request) {}

func (c *controller) UpdateCourier(w http.ResponseWriter, r *http.Request) {}
