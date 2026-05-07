package users

import (
	"context"
	"database/sql"
	"strings"
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

func (s *service) RegisterByTgID(ctx context.Context, params RegisterByTgID) error {
	createParams := CreateParams{
		TgID: sql.NullInt64{
			Int64: params.UserID,
			Valid: params.UserID > 0,
		},
		FullName: sql.NullString{
			String: params.Username,
			Valid:  strings.Trim(params.Username, " ") != "",
		},
		Role: RoleTypeCourier,
	}

	if err := s.usersRepository.Create(ctx, createParams); err != nil {
		return err
	}

	return nil
}

func (s *service) GetByTgID(ctx context.Context, userID int64) (User, error) {
	user, err := s.usersRepository.GetByTgID(ctx, sql.NullInt64{
		Int64: userID,
		Valid: userID > 0,
	})
	if err != nil {
		return User{}, err
	}
	return user, nil
}

type UpdatePhone struct {
	UserID int64
	Phone  string
}

func (s *service) UpdatePhone(ctx context.Context, params UpdatePhone) error {
	updatePhoneByTgIDParams := UpdatePhoneByTgIDParams{
		Phone: sql.NullString{
			String: params.Phone,
			Valid:  params.Phone != "",
		},
		TgID: sql.NullInt64{
			Int64: params.UserID,
			Valid: params.UserID > 0,
		},
	}

	return s.usersRepository.UpdatePhoneByTgID(ctx, updatePhoneByTgIDParams)
}
