package orders

import (
	"context"
	"database/sql"
	"strconv"
)

type Service interface {
	Create(ctx context.Context, order Create) (int64, error)
	GetByID(ctx context.Context, args GetByID) (*Order, error)
}

type service struct {
	ordersRepository Querier
}

func NewService(ordersRepository Querier) Service {
	return &service{
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
	Price          float64
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
		Price:          strconv.FormatFloat(order.Price, 'g', -1, 64),
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
