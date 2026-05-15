package utils

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"strconv"
)

const userIDKey = "user_id"

func NullStringToFloat(s sql.NullString) float64 {
	if !s.Valid || s.String == "" {
		return 0.0
	}

	val, err := strconv.ParseFloat(s.String, 64)
	if err != nil {
		return 0.0
	}
	return val
}

func GetFromCtx(ctx context.Context) (int64, error) {
	uid, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0, errors.New("user_id not found in context")
	}
	return uid, nil
}
