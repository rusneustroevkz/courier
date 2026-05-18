package utils

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
	"strconv"
)

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
	uid, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		return 0, fmt.Errorf("could not get user id")
	}
	return uid, nil
}
