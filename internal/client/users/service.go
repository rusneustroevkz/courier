package users

import (
	"context"

	"github.com/rusneustroevkz/courier/internal/client/telegram"
)

type Service interface {
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

func (s *service) GetByID(ctx context.Context, userID int64) (*User, error) {
	user, err := s.usersRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
