package organizations_branches

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rusneustroevkz/courier/internal/client/dadata"
)

type Service interface {
	AddAddress(ctx context.Context, address AddAddress) error
	List(ctx context.Context, organizationID int64) ([]*OrganizationBranch, error)
	GetByID(ctx context.Context, args GetByID) (*OrganizationBranch, error)
	Suggest(ctx context.Context, address, userAgent string) ([]dadata.SearchResults, error)
}

type service struct {
	cl         *http.Client
	dadataRepo dadata.Dadata
	repo       Querier
}

func NewService(dadataRepo dadata.Dadata, repo Querier) Service {
	return &service{
		cl: &http.Client{
			Timeout: 5 * time.Second,
		},
		dadataRepo: dadataRepo,
		repo:       repo,
	}
}

type AddAddress struct {
	ID             int64
	OrganizationID int64
	Name           string
	Address        string
	Latitude       float64
	Longitude      float64
	Phone          []string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *service) AddAddress(ctx context.Context, address AddAddress) error {
	getByNameParams := GetByNameParams{
		OrganizationID: address.OrganizationID,
		Name:           address.Name,
	}
	exists, err := s.repo.GetByName(ctx, getByNameParams)
	if err != nil {
		return err
	}

	if exists != nil {
		return fmt.Errorf("organization branch %s already exists", address.Name)
	}

	createParams := CreateParams{
		OrganizationID: address.OrganizationID,
		Name:           address.Name,
		Address:        address.Address,
		Latitude: sql.NullString{
			String: strconv.FormatFloat(address.Latitude, 'f', -1, 64),
			Valid:  address.Latitude > 0,
		},
		Longitude: sql.NullString{
			String: strconv.FormatFloat(address.Longitude, 'f', -1, 64),
			Valid:  address.Longitude > 0,
		},
		Phone: address.Phone,
	}

	if err := s.repo.Create(ctx, createParams); err != nil {
		return err
	}

	return nil
}

func (s *service) List(ctx context.Context, organizationID int64) ([]*OrganizationBranch, error) {
	res, err := s.repo.List(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type GetByID struct {
	OrganizationID int64
	BranchID       int64
}

func (s *service) GetByID(ctx context.Context, args GetByID) (*OrganizationBranch, error) {
	getByIDParams := GetByIDParams{
		OrganizationID: args.OrganizationID,
		ID:             args.BranchID,
	}

	return s.repo.GetByID(ctx, getByIDParams)
}

func (s *service) Suggest(ctx context.Context, address, userAgent string) ([]dadata.SearchResults, error) {
	return s.dadataRepo.Suggest(ctx, address, userAgent)
}
