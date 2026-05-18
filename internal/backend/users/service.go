package users

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type Service interface {
	RegisterByTgID(ctx context.Context, params RegisterByTgID) error
	GetByTgID(ctx context.Context, userID int64) (*GetByTgID, error)
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

type GetByTgID struct {
	ID        int64
	TgID      sql.NullInt64
	FullName  sql.NullString
	Email     sql.NullString
	Phone     sql.NullString
	Role      RoleType
	OnWork    bool
	Verified  bool
	Rating    sql.NullString
	Balance   sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *service) GetByTgID(ctx context.Context, userID int64) (*GetByTgID, error) {
	user, err := s.usersRepository.GetByTgID(ctx, sql.NullInt64{
		Int64: userID,
		Valid: userID > 0,
	})
	if err != nil {
		return nil, err
	}

	result := &GetByTgID{
		ID:        user.ID,
		TgID:      user.TgID,
		FullName:  user.FullName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		OnWork:    user.OnWork,
		Verified:  user.Verified,
		Rating:    user.Rating,
		Balance:   user.Balance,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return result, nil
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
