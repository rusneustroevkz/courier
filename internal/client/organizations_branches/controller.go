package organizations_branches

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
	List(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Suggest(w http.ResponseWriter, r *http.Request)
	SetActivation(w http.ResponseWriter, r *http.Request)
	SetUserSelected(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	service      Service
	usersService users.Service
}

func NewController(service Service, usersService users.Service) Controller {
	return &controller{
		service:      service,
		usersService: usersService,
	}
}

type CreateRequest struct {
	Name      string   `json:"name" validate:"required"`
	Address   string   `json:"address" validate:"required"`
	Latitude  float64  `json:"latitude" validate:"required,latitude"`
	Longitude float64  `json:"longitude" validate:"required,longitude"`
	Phone     []string `json:"phone" validate:"required,dive,e164"`
}
type CreateResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
	Data   *CreateData       `json:"data,omitempty"`
}
type CreateData struct {
	ID int64 `json:"id,omitempty"`
}

// Create Создание филиала
//
//	@Summary      Создание филиала
//	@Description  Создание филиала
//	@Tags         organizations-branches
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body CreateRequest true "тело запроса"
//	@Success      200  {object} CreateResponse
//	@Failure      400  {object} CreateResponse
//	@Failure      401  {object} CreateResponse
//	@Failure      404  {object} CreateResponse
//	@Failure      500  {object} CreateResponse
//	@Router       /organizations-branches [post]
func (c *controller) Create(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsBranchesCreate")

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
		Name:           req.Name,
		Address:        req.Address,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Phone:          req.Phone,
	}
	id, err := c.service.Create(r.Context(), params)
	if err != nil {
		log.ErrorContext(r.Context(), "failed create", "error", err)
		res.Errors["message"] = "Не удалось создать"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &CreateData{
		ID: id,
	}

	responder.Responder(w, res, http.StatusOK)
}

type ListRequest struct {
	ID          int64     `json:"id,omitempty"`
	CourierID   int64     `json:"courier_id,omitempty"`
	Status      string    `json:"status,omitempty"`
	FromAddress string    `json:"from_address,omitempty"`
	ToAddress   string    `json:"to_address,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Page        int64     `json:"page"`
	PageSize    int64     `json:"page_size"`
}
type ListResponse struct {
	Errors     map[string]string `json:"errors,omitempty"`
	Data       []ListData        `json:"data,omitempty"`
	Pagination *ListPagination   `json:"pagination,omitempty"`
}
type ListPagination struct {
	Page       int64 `json:"page,omitempty"`        // Текущая страница
	PageSize   int64 `json:"page_size,omitempty"`   // Размер страницы
	TotalCount int64 `json:"total_count,omitempty"` // Всего записей в БД по этому фильтру
	TotalPages int64 `json:"total_pages,omitempty"`
}
type ListData struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Phone          []string  `json:"phone"`
	IsActive       bool      `json:"is_active"`
	UserSelected   int64     `json:"user_selected"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// List Список
//
//	@Summary      Список
//	@Description  Список
//	@Tags         organizations-branches
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body ListRequest true "тело запроса"
//	@Success      200  {object} ListResponse
//	@Failure      400  {object} ListResponse
//	@Failure      401  {object} ListResponse
//	@Failure      404  {object} ListResponse
//	@Failure      500  {object} ListResponse
//	@Router       /organizations-branches/list [post]
func (c *controller) List(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsBranchesList")

	res := &ListResponse{
		Errors: make(map[string]string),
	}

	var req ListRequest
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

	params := List{
		OrganizationID: user.OrganizationID.Int64,
		Page:           req.Page,
		PageSize:       req.PageSize,
	}
	result, err := c.service.List(r.Context(), params)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get list", "error", err)
		res.Errors["message"] = "Ошибка списка"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Pagination = &ListPagination{
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	}
	for _, val := range result.Data {
		item := ListData{
			ID:             val.ID,
			OrganizationID: val.OrganizationID,
			Name:           val.Name,
			Address:        val.Address,
			Latitude:       val.Latitude,
			Longitude:      val.Longitude,
			Phone:          val.Phone,
			IsActive:       val.IsActive,
			CreatedAt:      val.CreatedAt,
			UpdatedAt:      val.UpdatedAt,
		}
		if val.UserSelected.Valid {
			item.UserSelected = val.UserSelected.Int64
		}

		res.Data = append(res.Data, item)
	}

	responder.Responder(w, res, http.StatusOK)
}

type GetByIDResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
	Data   *GetByIDData      `json:"data,omitempty"`
}
type GetByIDData struct {
	ID             int64     `json:"id"`
	OrganizationID int64     `json:"organization_id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Phone          []string  `json:"phone"`
	IsActive       bool      `json:"is_active"`
	UserSelected   int64     `json:"user_selected"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetByID Выборка
//
//	@Summary      Выборка
//	@Description  Выборка
//	@Tags         organizations-branches
//	@Accept       application/json
//	@Produce      application/json
//	@Param        id path int true "Идентификатор"
//	@Success      200  {object} GetByIDResponse
//	@Failure      400  {object} GetByIDResponse
//	@Failure      401  {object} GetByIDResponse
//	@Failure      404  {object} GetByIDResponse
//	@Failure      500  {object} GetByIDResponse
//	@Router       /organizations-branches/{id} [get]
func (c *controller) GetByID(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsBranchesGetByID")

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
	getByIDResult, err := c.service.GetByID(r.Context(), params)
	if err != nil {
		log.ErrorContext(r.Context(), "failed get", "error", err)
		res.Errors["message"] = "Ошибка выборки"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	res.Data = &GetByIDData{
		ID:             getByIDResult.ID,
		OrganizationID: getByIDResult.OrganizationID,
		Name:           getByIDResult.Name,
		Address:        getByIDResult.Address,
		Latitude:       getByIDResult.Latitude,
		Longitude:      getByIDResult.Longitude,
		Phone:          getByIDResult.Phone,
		IsActive:       getByIDResult.IsActive,
		CreatedAt:      getByIDResult.CreatedAt,
		UpdatedAt:      getByIDResult.UpdatedAt,
	}

	responder.Responder(w, res, http.StatusOK)
}

type SuggestRequest struct {
	Address string `json:"address" validate:"required,gt=5"`
}
type SuggestResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
	Data   []SuggestData     `json:"data,omitempty"`
}
type SuggestData struct {
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Address string `json:"address"`
}

// Suggest Подсказка адреса
//
//	@Summary      Подсказка адреса
//	@Description  Подсказка адреса
//	@Tags         organizations-branches
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body SuggestRequest true "тело запроса"
//	@Success      200  {object} SuggestResponse
//	@Failure      400  {object} SuggestResponse
//	@Failure      401  {object} SuggestResponse
//	@Failure      404  {object} SuggestResponse
//	@Failure      500  {object} SuggestResponse
//	@Router       /organizations-branches/suggest [post]
func (c *controller) Suggest(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsBranchesGetByID")

	res := &SuggestResponse{
		Errors: make(map[string]string),
	}

	var req SuggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.ErrorContext(r.Context(), "failed decode request", "error", err)
		res.Errors["message"] = "Ошибка обработки запроса"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusBadRequest)
		return
	}

	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	result, err := c.service.Suggest(r.Context(), req.Address, userAgent)
	if err != nil {
		log.ErrorContext(r.Context(), "failed suggest", "error", err)
		res.Errors["message"] = "Ошибка подсказки"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	for _, val := range result {
		res.Data = append(res.Data, SuggestData{
			Lat:     val.Lat,
			Lon:     val.Lon,
			Address: val.Address,
		})
	}

	responder.Responder(w, res, http.StatusOK)
}

type SetActivationRequest struct {
	ID       int64 `json:"id"`
	IsActive bool  `json:"is_active"`
}
type SetActivationResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
}

// SetActivation Изменить активный
//
//	@Summary      Изменить активный
//	@Description  Изменить активный
//	@Tags         organizations-branches
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body SetActivationRequest true "тело запроса"
//	@Success      200  {object} SetActivationResponse
//	@Failure      400  {object} SetActivationResponse
//	@Failure      401  {object} SetActivationResponse
//	@Failure      404  {object} SetActivationResponse
//	@Failure      500  {object} SetActivationResponse
//	@Router       /organizations-branches/set-activation [post]
func (c *controller) SetActivation(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsBranchesSetActivation")

	res := &SetActivationResponse{
		Errors: make(map[string]string),
	}

	var req SetActivationRequest
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

	params := SetActivation{
		IsActive:       req.IsActive,
		OrganizationID: user.OrganizationID.Int64,
		ID:             req.ID,
	}
	if err := c.service.SetActivation(r.Context(), params); err != nil {
		log.ErrorContext(r.Context(), "failed suggest", "error", err)
		res.Errors["message"] = "Ошибка подсказки"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	responder.Responder(w, res, http.StatusOK)
}

type SetUserSelectedRequest struct {
	ID int64 `json:"id"`
}
type SetUserSelectedResponse struct {
	Errors map[string]string `json:"errors,omitempty"`
}

// SetUserSelected Выбрать филиал
//
//	@Summary      Выбрать филиал
//	@Description  Выбрать филиал
//	@Tags         organizations-branches
//	@Accept       application/json
//	@Produce      application/json
//	@Param        request body SetUserSelectedRequest true "тело запроса"
//	@Success      200  {object} SetUserSelectedResponse
//	@Failure      400  {object} SetUserSelectedResponse
//	@Failure      401  {object} SetUserSelectedResponse
//	@Failure      404  {object} SetUserSelectedResponse
//	@Failure      500  {object} SetUserSelectedResponse
//	@Router       /organizations-branches/set-user-selected [post]
func (c *controller) SetUserSelected(w http.ResponseWriter, r *http.Request) {
	log := slog.With("method", "OrganizationsBranchesSetUserSelected")

	res := &SetUserSelectedResponse{
		Errors: make(map[string]string),
	}

	var req SetUserSelectedRequest
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

	params := SetUserSelected{
		OrganizationID: user.OrganizationID.Int64,
		ID:             req.ID,
		UserID:         userID,
	}
	if err := c.service.SetUserSelected(r.Context(), params); err != nil {
		log.ErrorContext(r.Context(), "failed suggest", "error", err)
		res.Errors["message"] = "Ошибка подсказки"
		res.Errors["error"] = err.Error()
		responder.Responder(w, res, http.StatusInternalServerError)
		return
	}

	responder.Responder(w, res, http.StatusOK)
}
