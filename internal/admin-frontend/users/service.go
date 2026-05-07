package users

import (
	"context"

	"github.com/rusneustroevkz/courier/internal/admin-frontend/telegram"
)

type Service interface {
	List(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, userID int64) (*User, error)
}

type service struct {
	usersRepository Querier
	telegramBot     *telegram.Telegram
}

func NewService(usersRepository Querier, telegramBot *telegram.Telegram) Service {
	return &service{
		usersRepository: usersRepository,
		telegramBot:     telegramBot,
	}
}

type RegisterByTgID struct {
	UserID   int64
	Username string
}

func (s *service) List(ctx context.Context) ([]User, error) {
	params := ListParams{
		Limit:  1000,
		Offset: 0,
	}
	return s.usersRepository.List(ctx, params)
}

func (s *service) GetByID(ctx context.Context, userID int64) (*User, error) {
	user, err := s.usersRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
