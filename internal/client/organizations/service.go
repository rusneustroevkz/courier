package organizations

import (
	"context"
)

type Service interface {
	My(ctx context.Context, userID int64) (*MyResult, error)
}

type service struct {
	organizationsRepo Querier
}

func NewService(organizationsRepo Querier) Service {
	return &service{
		organizationsRepo: organizationsRepo,
	}
}

type MyResult struct {
	ID      int64
	Name    string
	Balance string
}

func (s *service) My(ctx context.Context, userID int64) (*MyResult, error) {
	organization, err := s.organizationsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := &MyResult{
		ID:      organization.ID,
		Name:    organization.Name,
		Balance: organization.Balance,
	}

	return res, nil
}
