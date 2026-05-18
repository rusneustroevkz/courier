package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/client/config"
	"github.com/rusneustroevkz/courier/internal/client/users"
)

const (
	accessTokenTTL  = 60 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type Service interface {
	Register(ctx context.Context, args Register) error
	Login(ctx context.Context, email, password string) (*users.User, string, string, error)
	Refresh(ctx context.Context, oldRefreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	VerifyToken(tokenString string) (int64, error)
}

type service struct {
	usersRepository users.Querier
	authRepository  Querier
	cfg             *config.Config
}

func NewService(cfg *config.Config, usersRepository users.Querier, authRepository Querier) Service {
	return &service{
		usersRepository: usersRepository,
		authRepository:  authRepository,
		cfg:             cfg,
	}
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	err := s.authRepository.RevokeToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Refresh(ctx context.Context, oldRefreshToken string) (string, string, error) {
	storedToken, err := s.authRepository.GetActiveRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	accessTokenDuration := time.Now().Add(accessTokenTTL)
	accessToken, err := generateJWT(storedToken.UserID, accessTokenDuration, []byte(s.cfg.JWTAccessTokenSecret))
	if err != nil {
		return "", "", err
	}

	refreshTokenDuration := time.Now().Add(refreshTokenTTL)
	refreshToken, err := generateJWT(storedToken.UserID, refreshTokenDuration, []byte(s.cfg.JWTRefreshTokenSecret))
	if err != nil {
		return "", "", err
	}

	saveParams := SaveParams{
		Token:     refreshToken,
		ExpiresAt: refreshTokenDuration,
		UserID:    storedToken.UserID,
	}
	if err := s.authRepository.Save(ctx, saveParams); err != nil {
		return "", "", err
	}

	return refreshToken, accessToken, nil
}

type Register struct {
	Email    string
	Password string
	Fullname string
}

func (s *service) Register(ctx context.Context, args Register) error {
	exists, err := s.usersRepository.GetByEmail(ctx, sql.NullString{String: args.Email, Valid: args.Email != ""})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if exists != nil && exists.ID > 0 {
		return errors.New("Такой пользователь уже зарегистрирован")
	}

	hash, err := passwordHash(args.Password)
	if err != nil {
		return err
	}

	createByEmailParams := users.CreateByEmailParams{
		Email:        sql.NullString{String: args.Email, Valid: args.Email != ""},
		Role:         users.RoleTypeClient,
		PasswordHash: sql.NullString{String: hash, Valid: hash != ""},
		FullName:     sql.NullString{String: args.Fullname, Valid: args.Fullname != ""},
	}

	if !createByEmailParams.PasswordHash.Valid {
		return errors.New("invalid password hash")
	}
	if !createByEmailParams.Email.Valid {
		return errors.New("invalid email")
	}

	err = s.usersRepository.CreateByEmail(ctx, createByEmailParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Login(ctx context.Context, email, password string) (*users.User, string, string, error) {
	user, err := s.usersRepository.GetByEmail(ctx, sql.NullString{
		String: email,
		Valid:  email != "",
	})
	if err != nil {
		return nil, "", "", err
	}

	ok, err := comparePasswords(password, user.PasswordHash.String)
	if err != nil {
		return nil, "", "", err
	}
	if !ok {
		return nil, "", "", errors.New("passwords do not match")
	}

	if !user.Verified {
		return nil, "", "", errors.New("Пользователь еще не прошел проверку")
	}
	if !user.PasswordHash.Valid {
		return nil, "", "", errors.New("invalid password hash")
	}

	accessTokenDuration := time.Now().Add(accessTokenTTL)
	accessToken, err := generateJWT(user.ID, accessTokenDuration, []byte(s.cfg.JWTAccessTokenSecret))
	if err != nil {
		return nil, "", "", err
	}

	refreshTokenDuration := time.Now().Add(refreshTokenTTL)
	refreshToken, err := generateJWT(user.ID, refreshTokenDuration, []byte(s.cfg.JWTRefreshTokenSecret))
	if err != nil {
		return nil, "", "", err
	}

	saveParams := SaveParams{
		Token:     refreshToken,
		ExpiresAt: refreshTokenDuration,
		UserID:    user.ID,
	}
	if err := s.authRepository.Save(ctx, saveParams); err != nil {
		return nil, "", "", err
	}

	return user, refreshToken, accessToken, nil
}

func (s *service) VerifyToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTAccessTokenSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, errors.New("invalid token")
}

func comparePasswords(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func passwordHash(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func generateJWT(userID int64, refreshTokenDuration time.Time, secret []byte) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenDuration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
