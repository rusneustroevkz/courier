package users

import (
	"context"
	"database/sql"
	"time"
)

type Service interface {
	GetByID(ctx context.Context, userID int64) (*GetByIDResult, error)
}

type service struct {
	usersRepository Querier
}

func NewService(usersRepository Querier) Service {
	return &service{
		usersRepository: usersRepository,
	}
}

type GetByIDResult struct {
	ID             int64
	TgID           sql.NullInt64
	FullName       sql.NullString
	Email          sql.NullString
	Phone          sql.NullString
	Role           RoleType
	OnWork         bool
	Verified       bool
	Rating         sql.NullString
	Balance        sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PasswordHash   sql.NullString
	OrganizationID sql.NullInt64
}

func (s *service) GetByID(ctx context.Context, userID int64) (*GetByIDResult, error) {
	user, err := s.usersRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := &GetByIDResult{
		ID:             user.ID,
		TgID:           user.TgID,
		FullName:       user.FullName,
		Email:          user.Email,
		Phone:          user.Phone,
		Role:           user.Role,
		OnWork:         user.OnWork,
		Verified:       user.Verified,
		Rating:         user.Rating,
		Balance:        user.Balance,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		PasswordHash:   user.PasswordHash,
		OrganizationID: user.OrganizationID,
	}

	return res, nil
}
