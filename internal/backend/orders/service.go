package orders

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/rusneustroevkz/courier/internal/backend/organizations"
	"strconv"
	"time"
)

type Service interface {
	GetByID(ctx context.Context, orderID int64) (*GetByIDResult, error)
	GetPendingOrders(ctx context.Context) ([]GetByIDResult, error)
	AcceptOrder(ctx context.Context, args AcceptOrder) error
	GetActiveOrder(ctx context.Context, userID int64) (*GetByIDResult, error)
	DoneOrder(ctx context.Context, orderID int64) error
	PickUpOrder(ctx context.Context, orderID int64) error
}

type service struct {
	ordersRepository        *Queries
	organizationsRepository *organizations.Queries
	db                      *sqlx.DB
}

func NewService(db *sqlx.DB, ordersRepository *Queries, organizationsRepository *organizations.Queries) Service {
	return &service{
		ordersRepository:        ordersRepository,
		organizationsRepository: organizationsRepository,
		db:                      db,
	}
}

type GetByIDResult struct {
	ID                     int64
	Description            string
	OrganizationID         int64
	CourierID              int64
	Status                 OrderStatus
	FromAddress            string
	FromLat                string
	FromLon                string
	ToAddress              string
	ToLat                  string
	ToLon                  string
	TgClientChatID         int64
	TgLiveMessageID        int64
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
}

func (s *service) GetByID(ctx context.Context, orderID int64) (*GetByIDResult, error) {
	getByIDResult, err := s.ordersRepository.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	result := GetByIDResult{
		ID:                     getByIDResult.ID,
		OrganizationID:         getByIDResult.OrganizationID,
		Status:                 getByIDResult.Status,
		FromAddress:            getByIDResult.FromAddress,
		FromLat:                getByIDResult.FromLat,
		FromLon:                getByIDResult.FromLon,
		ToAddress:              getByIDResult.ToAddress,
		ToLat:                  getByIDResult.ToLat,
		ToLon:                  getByIDResult.ToLon,
		CreatedAt:              getByIDResult.CreatedAt,
		UpdatedAt:              getByIDResult.UpdatedAt,
		CourierEarnings:        getByIDResult.CourierEarnings,
		DeliveryDistanceMeters: getByIDResult.DeliveryDistanceMeters,
	}

	if getByIDResult.Description.Valid {
		result.Description = getByIDResult.Description.String
	}
	if getByIDResult.CourierID.Valid {
		result.CourierID = getByIDResult.CourierID.Int64
	}
	if getByIDResult.BranchID.Valid {
		result.BranchID = getByIDResult.BranchID.Int64
	}
	if getByIDResult.TgCourierChatID.Valid {
		result.TgCourierChatID = getByIDResult.TgCourierChatID.Int64
	}
	if getByIDResult.AcceptedAt.Valid {
		result.AcceptedAt = getByIDResult.AcceptedAt.Time
	}
	if getByIDResult.PickedUpAt.Valid {
		result.PickedUpAt = getByIDResult.PickedUpAt.Time
	}
	if getByIDResult.DeliveredAt.Valid {
		result.DeliveredAt = getByIDResult.DeliveredAt.Time
	}
	if getByIDResult.CancelledAt.Valid {
		result.CancelledAt = getByIDResult.CancelledAt.Time
	}

	return &result, nil
}

func (s *service) GetPendingOrders(ctx context.Context) ([]GetByIDResult, error) {
	getPendingOrdersResult, err := s.ordersRepository.GetPendingOrders(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]GetByIDResult, len(getPendingOrdersResult))
	for i, item := range getPendingOrdersResult {
		val := GetByIDResult{
			ID:                     item.ID,
			OrganizationID:         item.OrganizationID,
			Status:                 item.Status,
			FromAddress:            item.FromAddress,
			FromLat:                item.FromLat,
			FromLon:                item.FromLon,
			ToAddress:              item.ToAddress,
			ToLat:                  item.ToLat,
			ToLon:                  item.ToLon,
			CreatedAt:              item.CreatedAt,
			UpdatedAt:              item.UpdatedAt,
			CourierEarnings:        item.CourierEarnings,
			DeliveryDistanceMeters: item.DeliveryDistanceMeters,
		}

		if item.Description.Valid {
			val.Description = item.Description.String
		}
		if item.CourierID.Valid {
			val.CourierID = item.CourierID.Int64
		}
		if item.BranchID.Valid {
			val.BranchID = item.BranchID.Int64
		}
		if item.TgCourierChatID.Valid {
			val.TgCourierChatID = item.TgCourierChatID.Int64
		}
		if item.AcceptedAt.Valid {
			val.AcceptedAt = item.AcceptedAt.Time
		}
		if item.PickedUpAt.Valid {
			val.PickedUpAt = item.PickedUpAt.Time
		}
		if item.DeliveredAt.Valid {
			val.DeliveredAt = item.DeliveredAt.Time
		}
		if item.CancelledAt.Valid {
			val.CancelledAt = item.CancelledAt.Time
		}

		result[i] = val
	}

	return result, nil
}

type AcceptOrder struct {
	CourierID int64
	Status    OrderStatus
	ID        int64
}

func (s *service) AcceptOrder(ctx context.Context, args AcceptOrder) error {
	acceptOrderParams := AcceptOrderParams{
		CourierID: sql.NullInt64{
			Int64: args.CourierID,
			Valid: args.CourierID != 0,
		},
		Status: args.Status,
		ID:     args.ID,
	}

	return s.ordersRepository.AcceptOrder(ctx, acceptOrderParams)
}

func (s *service) GetActiveOrder(ctx context.Context, userID int64) (*GetByIDResult, error) {
	getByIDResult, err := s.ordersRepository.GetCourierActiveOrder(ctx, sql.NullInt64{
		Int64: userID,
		Valid: userID != 0,
	})
	if err != nil {
		return nil, err
	}

	result := GetByIDResult{
		ID:                     getByIDResult.ID,
		OrganizationID:         getByIDResult.OrganizationID,
		Status:                 getByIDResult.Status,
		FromAddress:            getByIDResult.FromAddress,
		FromLat:                getByIDResult.FromLat,
		FromLon:                getByIDResult.FromLon,
		ToAddress:              getByIDResult.ToAddress,
		ToLat:                  getByIDResult.ToLat,
		ToLon:                  getByIDResult.ToLon,
		CreatedAt:              getByIDResult.CreatedAt,
		UpdatedAt:              getByIDResult.UpdatedAt,
		CourierEarnings:        getByIDResult.CourierEarnings,
		DeliveryDistanceMeters: getByIDResult.DeliveryDistanceMeters,
	}

	if getByIDResult.Description.Valid {
		result.Description = getByIDResult.Description.String
	}
	if getByIDResult.CourierID.Valid {
		result.CourierID = getByIDResult.CourierID.Int64
	}
	if getByIDResult.BranchID.Valid {
		result.BranchID = getByIDResult.BranchID.Int64
	}
	if getByIDResult.TgCourierChatID.Valid {
		result.TgCourierChatID = getByIDResult.TgCourierChatID.Int64
	}
	if getByIDResult.AcceptedAt.Valid {
		result.AcceptedAt = getByIDResult.AcceptedAt.Time
	}
	if getByIDResult.PickedUpAt.Valid {
		result.PickedUpAt = getByIDResult.PickedUpAt.Time
	}
	if getByIDResult.DeliveredAt.Valid {
		result.DeliveredAt = getByIDResult.DeliveredAt.Time
	}
	if getByIDResult.CancelledAt.Valid {
		result.CancelledAt = getByIDResult.CancelledAt.Time
	}

	return &result, nil
}

func (s *service) DoneOrder(ctx context.Context, orderID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	order, err := s.ordersRepository.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := s.ordersRepository.WithTx(tx).DoneOrder(ctx, orderID); err != nil {
		return err
	}

	organization, err := s.organizationsRepository.GetByID(ctx, order.OrganizationID)
	if err != nil {
		return err
	}

	balance, err := strconv.ParseFloat(organization.Balance, 64)
	if err != nil {
		return err
	}
	newBalance := balance - 100.00

	payOrderParams := organizations.PayOrderParams{
		Balance: strconv.FormatFloat(newBalance, 'f', -1, 64),
		ID:      organization.ID,
	}
	if err := s.organizationsRepository.WithTx(tx).PayOrder(ctx, payOrderParams); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *service) PickUpOrder(ctx context.Context, orderID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	order, err := s.ordersRepository.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := s.ordersRepository.WithTx(tx).PickUpOrder(ctx, orderID); err != nil {
		return err
	}

	organization, err := s.organizationsRepository.GetByID(ctx, order.OrganizationID)
	if err != nil {
		return err
	}

	balance, err := strconv.ParseFloat(organization.Balance, 64)
	if err != nil {
		return err
	}
	newBalance := balance - 100.00

	payOrderParams := organizations.PayOrderParams{
		Balance: strconv.FormatFloat(newBalance, 'f', -1, 64),
		ID:      organization.ID,
	}
	if err := s.organizationsRepository.WithTx(tx).PayOrder(ctx, payOrderParams); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
