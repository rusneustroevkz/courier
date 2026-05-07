package users

import (
	"context"
)

type Service interface {
	RegisterByTgID(ctx context.Context, params RegisterByTgID) error
	GetByTgID(ctx context.Context, userID int64) (User, error)
	UpdatePhone(ctx context.Context, params UpdatePhone) error
}

type service struct {
	usersRepository Querier
}

func NewService(usersRepository Querier) Service {
	return &service{
		usersRepository: usersRepository,
	}
}

type RegisterByTgID struct {
	UserID   int64
	Username string
}

func (s *service) List(ctx context.Context) ([]User, error) {

}
