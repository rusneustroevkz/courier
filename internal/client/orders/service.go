package orders

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, order Create) (int64, error)
	GetByID(ctx context.Context, args GetByID) (*Order, error)
	GetAll(ctx context.Context, filter GetAll) (*GetAllResult, error)
}

type service struct {
	db               *sqlx.DB
	ordersRepository Querier
}

func NewService(db *sqlx.DB, ordersRepository Querier) Service {
	return &service{
		db:               db,
		ordersRepository: ordersRepository,
	}
}

type Create struct {
	OrganizationID int64
	FromAddress    string
	FromLat        float64
	FromLon        float64
	ToAddress      string
	ToLat          float64
	ToLon          float64
	Description    string
}

func (s *service) Create(ctx context.Context, order Create) (int64, error) {
	params := CreateOrderParams{
		OrganizationID: order.OrganizationID,
		FromAddress:    order.FromAddress,
		FromLat:        strconv.FormatFloat(order.FromLat, 'g', -1, 64),
		FromLon:        strconv.FormatFloat(order.FromLon, 'g', -1, 64),
		ToAddress:      order.ToAddress,
		ToLat:          strconv.FormatFloat(order.ToLat, 'g', -1, 64),
		ToLon:          strconv.FormatFloat(order.ToLon, 'g', -1, 64),
		Description: sql.NullString{
			String: order.Description,
			Valid:  order.Description != "",
		},
	}

	return s.ordersRepository.CreateOrder(ctx, params)
}

type GetByID struct {
	ID             int64
	OrganizationID int64
}

func (s *service) GetByID(ctx context.Context, args GetByID) (*Order, error) {
	params := GetByIDParams{
		ID:             args.ID,
		OrganizationID: args.OrganizationID,
	}
	return s.ordersRepository.GetByID(ctx, params)
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
	TotalCount int64 `json:"total_count"` // Всего записей в БД по этому фильтру
	TotalPages int64 `json:"total_pages"`
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
