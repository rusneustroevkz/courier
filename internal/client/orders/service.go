package orders

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	redislib "github.com/redis/go-redis/v9"
	"github.com/rusneustroevkz/courier/internal/client/organizations_branches"
	"github.com/rusneustroevkz/courier/internal/client/telegram"
	backendTelegram "github.com/rusneustroevkz/courier/internal/client/telegram"
	"github.com/rusneustroevkz/courier/internal/client/users"
	"github.com/rusneustroevkz/courier/pkg/redis"
	"gopkg.in/telebot.v4"
)

type Service interface {
	Create(ctx context.Context, order Create) (int64, error)
	GetByID(ctx context.Context, args GetByID) (*GetByIDResult, error)
	GetAll(ctx context.Context, filter GetAll) (*GetAllResult, error)
	CancelOrder(ctx context.Context, args CancelOrder) error
	Update(ctx context.Context, args Update) error
}

type service struct {
	db               *sqlx.DB
	ordersRepository Querier
	branchesRepo     organizations_branches.Querier
	redisClient      *redis.Redis
	telegramBot      *telegram.Telegram
	usersRepo        users.Querier
}

func NewService(
	db *sqlx.DB,
	ordersRepository Querier,
	branchesRepo organizations_branches.Querier,
	redisClient *redis.Redis,
	telegramBot *telegram.Telegram,
	usersRepo users.Querier,
) Service {
	return &service{
		db:               db,
		ordersRepository: ordersRepository,
		branchesRepo:     branchesRepo,
		redisClient:      redisClient,
		telegramBot:      telegramBot,
		usersRepo:        usersRepo,
	}
}

type Create struct {
	OrganizationID int64
	UserID         int64
	ToAddress      string
	ToLat          float64
	ToLon          float64
	Description    string
}

func (s *service) Create(ctx context.Context, args Create) (int64, error) {
	getCurrentSelectedParams := organizations_branches.GetCurrentSelectedParams{
		OrganizationID: args.OrganizationID,
		UserSelected:   sql.NullInt64{Int64: args.UserID, Valid: args.UserID != 0},
	}
	branch, err := s.branchesRepo.GetCurrentSelected(ctx, getCurrentSelectedParams)
	if err != nil {
		return 0, err
	}

	if branch == nil || branch.ID == 0 {
		return 0, errors.New("branch not found")
	}
	if !branch.Latitude.Valid || !branch.Longitude.Valid {
		return 0, errors.New("invalid coords")
	}
	if branch.Latitude.String == "" || branch.Longitude.String == "" {
		return 0, errors.New("bad coords")
	}

	params := CreateOrderParams{
		OrganizationID: args.OrganizationID,
		FromAddress:    branch.Address,
		FromLat:        branch.Latitude.String,
		FromLon:        branch.Longitude.String,
		ToAddress:      args.ToAddress,
		ToLat:          strconv.FormatFloat(args.ToLat, 'g', -1, 64),
		ToLon:          strconv.FormatFloat(args.ToLon, 'g', -1, 64),
		Description: sql.NullString{
			String: args.Description,
			Valid:  args.Description != "",
		},
		Price: sql.NullString{
			String: strconv.FormatFloat(100.00, 'g', -1, 64),
			Valid:  true,
		},
	}

	id, err := s.ordersRepository.CreateOrder(ctx, params)
	if err != nil {
		return 0, err
	}
	idString := strconv.FormatInt(id, 10)

	couriersOnWork, err := s.usersRepo.ListOnWorkCourier(ctx)
	if err != nil {
		return 0, err
	}

	text := fmt.Sprintf("Новый заказ\n\nОт: %s\nДо: %s", branch.Address, args.ToAddress)

	if args.Description != "" {
		text += "\nОписание: " + args.Description
	}

	selector := &telebot.ReplyMarkup{}
	rows := []telebot.Row{
		{selector.Data("Принять", backendTelegram.CallbackAcceptOrder, idString)},
	}

	selector.Inline(rows...)

	for _, val := range couriersOnWork {
		if !val.TgID.Valid {
			slog.ErrorContext(ctx, "invalid tg id to send message on create order")
			continue
		}
		if err := s.telegramBot.Send(ctx, val.TgID.Int64, text, selector); err != nil {
			slog.ErrorContext(ctx, "failed to send telegram msg on create order", "err", err, "user_id", val.TgID.Int64)
		}
	}

	return id, nil
}

type Update struct {
	OrderID        int64
	OrganizationID int64
	UserID         int64
	ToAddress      *string
	ToLat          *float64
	ToLon          *float64
	Description    *string
}

func (s *service) Update(ctx context.Context, args Update) error {
	getCurrentSelectedParams := organizations_branches.GetCurrentSelectedParams{
		OrganizationID: args.OrganizationID,
		UserSelected:   sql.NullInt64{Int64: args.UserID, Valid: args.UserID != 0},
	}
	branch, err := s.branchesRepo.GetCurrentSelected(ctx, getCurrentSelectedParams)
	if err != nil {
		return err
	}

	if branch == nil || branch.ID == 0 {
		return errors.New("branch not found")
	}
	if !branch.Latitude.Valid || !branch.Longitude.Valid {
		return errors.New("invalid coords")
	}
	if branch.Latitude.String == "" || branch.Longitude.String == "" {
		return errors.New("bad coords")
	}

	getByIDParams := GetByIDParams{
		ID:             args.OrderID,
		OrganizationID: args.OrganizationID,
	}
	order, err := s.ordersRepository.GetByID(ctx, getByIDParams)
	if err != nil {
		return err
	}

	params := UpdateParams{
		ToAddress:      order.ToAddress,
		ToLat:          order.ToLat,
		ToLon:          order.ToLon,
		Description:    order.Description,
		ID:             args.OrderID,
		OrganizationID: args.OrganizationID,
	}
	if args.ToAddress != nil {
		params.ToAddress = *args.ToAddress
		if args.ToLat == nil || args.ToLon == nil {
			return errors.New("bad coords")
		}
		params.ToLat = strconv.FormatFloat(*args.ToLat, 'g', -1, 64)
		params.ToLon = strconv.FormatFloat(*args.ToLon, 'g', -1, 64)
	}
	if args.Description != nil {
		params.Description = sql.NullString{
			String: *args.Description,
			Valid:  *args.Description != "",
		}
	}

	if err := s.ordersRepository.Update(ctx, params); err != nil {
		return err
	}

	toAddress := order.ToAddress
	if args.ToAddress != nil {
		toAddress = *args.ToAddress
	}

	text := fmt.Sprintf("Заказ изменился\n\nОт: %s\nДо: %s", branch.Address, toAddress)

	if args.Description != nil {
		text += "\nОписание: " + *args.Description
	}

	if order.CourierID.Valid {
		if err := s.telegramBot.Send(ctx, order.CourierID.Int64, text); err != nil {
			slog.ErrorContext(ctx, "failed to send telegram msg on create order", "err", err, "user_id", order.CourierID.Int64)
		}
	}

	return nil
}

type GetByID struct {
	ID             int64
	OrganizationID int64
}
type GetByIDResult struct {
	ID                     int64
	Description            string
	OrganizationID         int64
	Status                 string
	FromAddress            string
	FromLat                string
	FromLon                string
	ToAddress              string
	ToLat                  string
	ToLon                  string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	BranchID               int64
	CourierEarnings        string
	DeliveryDistanceMeters int32
	TgCourierChatID        int64
	AcceptedAt             time.Time
	PickedUpAt             time.Time
	DeliveredAt            time.Time
	CancelledAt            time.Time
	CourierID              int64
	CourierLat             float32
	CourierLon             float32
}

func (s *service) GetByID(ctx context.Context, args GetByID) (*GetByIDResult, error) {
	params := GetByIDParams{
		ID:             args.ID,
		OrganizationID: args.OrganizationID,
	}

	order, err := s.ordersRepository.GetByID(ctx, params)
	if err != nil {
		return nil, err
	}

	result := GetByIDResult{
		ID:                     order.ID,
		OrganizationID:         order.OrganizationID,
		Status:                 string(order.Status),
		FromAddress:            order.FromAddress,
		FromLat:                order.FromLat,
		FromLon:                order.FromLon,
		ToAddress:              order.ToAddress,
		ToLat:                  order.ToLat,
		ToLon:                  order.ToLon,
		CreatedAt:              order.CreatedAt,
		UpdatedAt:              order.UpdatedAt,
		CourierEarnings:        order.CourierEarnings,
		DeliveryDistanceMeters: order.DeliveryDistanceMeters,
	}

	if order.Description.Valid {
		result.Description = order.Description.String
	}
	if order.BranchID.Valid {
		result.BranchID = order.BranchID.Int64
	}
	if order.TgCourierChatID.Valid {
		result.TgCourierChatID = order.TgCourierChatID.Int64
	}
	if order.AcceptedAt.Valid {
		result.AcceptedAt = order.AcceptedAt.Time
	}
	if order.PickedUpAt.Valid {
		result.PickedUpAt = order.PickedUpAt.Time
	}
	if order.DeliveredAt.Valid {
		result.DeliveredAt = order.DeliveredAt.Time
	}
	if order.CancelledAt.Valid {
		result.CancelledAt = order.CancelledAt.Time
	}

	var courierIDString string

	if order.CourierID.Valid {
		result.CourierID = order.CourierID.Int64
		courierIDString = fmt.Sprintf("%d", order.CourierID.Int64)
	}

	orderIsActive := order.Status == OrderStatusAccepted || order.Status == OrderStatusPickedUp
	if courierIDString != "" && orderIsActive {
		val, err := s.redisClient.Client.Get(ctx, "location_"+courierIDString).Result()
		if errors.Is(err, redislib.Nil) {
			return &result, errors.New("Локация курьера не существует или срок его действия истек")
		} else if err != nil {
			return &result, errors.New("Ошибка при получении данных: " + err.Error())
		}

		var courierLocation redis.UserLocation

		err = json.Unmarshal([]byte(val), &courierLocation)
		if err != nil {
			return &result, errors.New("Ошибка декодирования данных: " + err.Error())
		}

		result.CourierLat = courierLocation.Latitude
		result.CourierLon = courierLocation.Longitude
	}

	return &result, nil
}

type GetAll struct {
	ID             sql.NullInt64
	OrganizationID sql.NullInt64
	CourierID      sql.NullInt64
	Status         sql.NullString
	FromAddress    sql.NullString
	ToAddress      sql.NullString
	CreatedAt      sql.NullTime
	UpdatedAt      sql.NullTime
	Page           int64
	PageSize       int64
}
type GetAllResult struct {
	Data       []Order
	TotalCount int64
	TotalPages int64
}

func (s *service) GetAll(ctx context.Context, filter GetAll) (*GetAllResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	baseQb := squirrel.Select().From("orders").PlaceholderFormat(squirrel.Dollar)

	if filter.ID.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"id": filter.ID.Int64})
	}
	if filter.OrganizationID.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"organization_id": filter.OrganizationID.Int64})
	}
	if filter.CourierID.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"courier_id": filter.CourierID.Int64})
	}
	if filter.Status.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"status": filter.Status.String})
	}
	if filter.FromAddress.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"from_address": filter.FromAddress.String})
	}
	if filter.ToAddress.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"to_address": filter.ToAddress.String})
	}
	if filter.CreatedAt.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"created_at": filter.CreatedAt.Time})
	}
	if filter.UpdatedAt.Valid {
		baseQb = baseQb.Where(squirrel.Eq{"updated_at": filter.UpdatedAt.Time})
	}

	offset := uint64((filter.Page - 1) * filter.PageSize)
	dataQb := baseQb.Columns("*").Limit(uint64(filter.PageSize)).Offset(offset)

	query, args, err := dataQb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql data query")
	}

	var orders []Order
	err = s.db.SelectContext(ctx, &orders, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query orders")
	}

	countQb := baseQb.Columns("COUNT(*)")

	countQuery, args, err := countQb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql count query")
	}

	var totalCount int64
	err = s.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query total count")
	}

	totalPages := totalCount / filter.PageSize
	if totalCount%filter.PageSize > 0 {
		totalPages++
	}

	return &GetAllResult{
		Data:       orders,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

type CancelOrder struct {
	OrderID        int64
	OrganizationID int64
}

func (s *service) CancelOrder(ctx context.Context, args CancelOrder) error {
	getByIDParams := GetByIDParams{
		ID:             args.OrderID,
		OrganizationID: args.OrganizationID,
	}
	order, err := s.ordersRepository.GetByID(ctx, getByIDParams)
	if err != nil {
		return err
	}

	cancelOrderParams := CancelOrderParams{
		Status:         OrderStatusCancelled,
		ID:             args.OrderID,
		OrganizationID: args.OrganizationID,
	}

	if err := s.ordersRepository.CancelOrder(ctx, cancelOrderParams); err != nil {
		return err
	}

	if order.CourierID.Valid {
		return s.telegramBot.Send(ctx, order.CourierID.Int64, "Заказ отменен")
	}

	return nil
}
