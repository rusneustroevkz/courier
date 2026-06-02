package organizations_branches

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"net/http"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/client/dadata"
)

type Service interface {
	Create(ctx context.Context, address Create) (int64, error)
	List(ctx context.Context, filter List) (*ListResult, error)
	GetByID(ctx context.Context, args GetByID) (*GetByIDResult, error)
	Suggest(ctx context.Context, address, userAgent string) ([]dadata.SearchResults, error)
	SetActivation(ctx context.Context, args SetActivation) error
	SetUserSelected(ctx context.Context, args SetUserSelected) error
}

type service struct {
	cl         *http.Client
	dadataRepo dadata.Dadata
	repo       *Queries
	db         *sqlx.DB
}

func NewService(dadataRepo dadata.Dadata, repo *Queries, db *sqlx.DB) Service {
	return &service{
		cl: &http.Client{
			Timeout: 5 * time.Second,
		},
		dadataRepo: dadataRepo,
		repo:       repo,
		db:         db,
	}
}

type Create struct {
	OrganizationID int64
	Name           string
	Address        string
	Latitude       float64
	Longitude      float64
	Phone          []string
}

func (s *service) Create(ctx context.Context, address Create) (int64, error) {
	getByNameParams := GetByNameParams{
		OrganizationID: address.OrganizationID,
		Name:           address.Name,
	}
	exists, err := s.repo.GetByName(ctx, getByNameParams)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if exists != nil && exists.ID > 0 {
		return 0, fmt.Errorf("organization branch %s already exists", address.Name)
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

	id, err := s.repo.Create(ctx, createParams)
	if err != nil {
		return 0, err
	}

	return id, nil
}

type List struct {
	OrganizationID int64
	Page           int64
	PageSize       int64
}
type ListResult struct {
	Data       []GetByIDResult
	TotalCount int64
	TotalPages int64
}

func (s *service) List(ctx context.Context, filter List) (*ListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	baseQb := squirrel.Select().From("organization_branches").PlaceholderFormat(squirrel.Dollar)

	baseQb = baseQb.Where(squirrel.Eq{"organization_id": filter.OrganizationID})

	offset := uint64((filter.Page - 1) * filter.PageSize)
	dataQb := baseQb.Columns("*").Limit(uint64(filter.PageSize)).Offset(offset)

	query, args, err := dataQb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql data query")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query orders")
	}

	defer rows.Close()

	var result []GetByIDResult

	var latitude sql.NullFloat64
	var longitude sql.NullFloat64

	for rows.Next() {
		var item GetByIDResult
		if err := rows.Scan(
			&item.ID,
			&item.OrganizationID,
			&item.Name,
			&item.Address,
			&latitude,
			&longitude,
			pq.Array(&item.Phone),
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.UserSelected,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		if latitude.Valid {
			item.Latitude = latitude.Float64
		}
		if longitude.Valid {
			item.Longitude = longitude.Float64
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	countQb := baseQb.Columns("COUNT(*)")

	countQuery, args, err := countQb.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql count query")
	}

	var totalCount int64
	err = s.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query total count")
	}

	totalPages := totalCount / filter.PageSize
	if totalCount%filter.PageSize > 0 {
		totalPages++
	}

	return &ListResult{
		Data:       result,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

type GetByID struct {
	OrganizationID int64
	ID             int64
}
type GetByIDResult struct {
	ID             int64
	OrganizationID int64
	Name           string
	Address        string
	Latitude       float64
	Longitude      float64
	Phone          []string
	IsActive       bool
	UserSelected   sql.NullInt64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *service) GetByID(ctx context.Context, args GetByID) (*GetByIDResult, error) {
	getByIDParams := GetByIDParams{
		OrganizationID: args.OrganizationID,
		ID:             args.ID,
	}

	item, err := s.repo.GetByID(ctx, getByIDParams)
	if err != nil {
		return nil, err
	}

	result := &GetByIDResult{
		ID:             item.ID,
		OrganizationID: item.OrganizationID,
		Name:           item.Name,
		Address:        item.Address,
		Phone:          item.Phone,
		IsActive:       item.IsActive,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
		UserSelected:   item.UserSelected,
	}
	if item.Latitude.Valid {
		result.Latitude, err = strconv.ParseFloat(item.Latitude.String, 64)
		if err != nil {
			return nil, err
		}
	}
	if item.Longitude.Valid {
		result.Longitude, err = strconv.ParseFloat(item.Longitude.String, 64)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *service) Suggest(ctx context.Context, address, userAgent string) ([]dadata.SearchResults, error) {
	return s.dadataRepo.Suggest(ctx, address, userAgent)
}

type SetActivation struct {
	IsActive       bool
	OrganizationID int64
	ID             int64
}

func (s *service) SetActivation(ctx context.Context, args SetActivation) error {
	params := SetActivationParams{
		IsActive:       args.IsActive,
		OrganizationID: args.OrganizationID,
		ID:             args.ID,
	}
	return s.repo.SetActivation(ctx, params)
}

type SetUserSelected struct {
	UserID         int64
	OrganizationID int64
	ID             int64
}

func (s *service) SetUserSelected(ctx context.Context, args SetUserSelected) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	getCurrentSelectedParams := GetCurrentSelectedParams{
		OrganizationID: args.OrganizationID,
		UserSelected:   sql.NullInt64{Valid: args.UserID > 0, Int64: args.UserID},
	}
	var branch *OrganizationBranch
	branch, err = s.repo.WithTx(tx).GetCurrentSelected(ctx, getCurrentSelectedParams)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if branch != nil && branch.ID > 0 {
		setNullUserSelectedParams := SetNullUserSelectedParams{
			OrganizationID: args.OrganizationID,
			ID:             branch.ID,
		}
		err = s.repo.WithTx(tx).SetNullUserSelected(ctx, setNullUserSelectedParams)
		if err != nil {
			return err
		}
	}

	setUserSelectedParams := SetUserSelectedParams{
		UserSelected: sql.NullInt64{
			Int64: args.UserID,
			Valid: args.UserID > 0,
		},
		OrganizationID: args.OrganizationID,
		ID:             args.ID,
	}
	err = s.repo.WithTx(tx).SetUserSelected(ctx, setUserSelectedParams)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
