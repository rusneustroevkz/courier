package users

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type Service interface {
	RegisterByTgID(ctx context.Context, params RegisterByTgID) (*GetByTgID, error)
	GetByTgID(ctx context.Context, userID int64) (*GetByTgID, error)
	UpdatePhone(ctx context.Context, params UpdatePhone) error
	SetOnWork(ctx context.Context, args SetOnWork) error
	SetShareLocation(ctx context.Context, args SetShareLocation) error
	WorkerSetShareLocationAfterTTL(ctx context.Context) error
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

func (s *service) RegisterByTgID(ctx context.Context, params RegisterByTgID) (*GetByTgID, error) {
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

	id, err := s.usersRepository.Create(ctx, createParams)
	if err != nil {
		return nil, err
	}

	result := &GetByTgID{
		ID:              id,
		OnWork:          false,
		Verified:        false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsShareLocation: false,
		Role:            string(RoleTypeCourier),
		TgID:            params.UserID,
		FullName:        params.Username,
		Rating:          "5.0",
		Balance:         "0.0",
	}

	return result, nil
}

type GetByTgID struct {
	ID              int64
	TgID            int64
	FullName        string
	Email           string
	Phone           string
	Role            string
	OnWork          bool
	Verified        bool
	Rating          string
	Balance         string
	IsShareLocation bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
		ID:              user.ID,
		OnWork:          user.OnWork,
		Verified:        user.Verified,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		IsShareLocation: user.IsShareLocation,
		Role:            string(user.Role),
	}

	if user.TgID.Valid {
		result.TgID = user.TgID.Int64
	}
	if user.FullName.Valid {
		result.FullName = user.FullName.String
	}
	if user.Email.Valid {
		result.Email = user.Email.String
	}
	if user.Phone.Valid {
		result.Phone = user.Phone.String
	}
	if user.Rating.Valid {
		result.Rating = user.Rating.String
	}
	if user.Balance.Valid {
		result.Balance = user.Balance.String
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

type SetOnWork struct {
	UserID int64
	OnWork bool
}

func (s *service) SetOnWork(ctx context.Context, args SetOnWork) error {
	setOnWorkParams := SetOnWorkParams{
		OnWork: args.OnWork,
		TgID: sql.NullInt64{
			Int64: args.UserID,
			Valid: args.UserID > 0,
		},
	}

	return s.usersRepository.SetOnWork(ctx, setOnWorkParams)
}

type SetShareLocation struct {
	TgUserID        int64
	IsShareLocation bool
	LivePeriod      time.Time
	OnWork          bool
}

func (s *service) SetShareLocation(ctx context.Context, args SetShareLocation) error {
	params := SetShareLocationParams{
		IsShareLocation: args.IsShareLocation,
		TgID: sql.NullInt64{
			Int64: args.TgUserID,
			Valid: args.TgUserID > 0,
		},
		ShareLocationTtl: sql.NullTime{
			Time:  args.LivePeriod,
			Valid: !args.LivePeriod.IsZero(),
		},
		OnWork: args.OnWork,
	}
	return s.usersRepository.SetShareLocation(ctx, params)
}

func (s *service) WorkerSetShareLocationAfterTTL(ctx context.Context) error {
	rows, err := s.usersRepository.ExpiredShareLocationList(ctx)
	if err != nil {
		return err
	}

	for _, val := range rows {
		if !val.ShareLocationTtl.Valid {
			continue
		}

		if val.ShareLocationTtl.Time.Before(time.Now()) {
			params := SetShareLocationParams{
				IsShareLocation: false,
				ShareLocationTtl: sql.NullTime{
					Time:  time.Now().Add(-1),
					Valid: true,
				},
				TgID: sql.NullInt64{
					Int64: val.ID,
					Valid: val.ID > 0,
				},
			}
			if err := s.usersRepository.SetShareLocation(ctx, params); err != nil {
				return err
			}
		}
	}

	return nil
}
