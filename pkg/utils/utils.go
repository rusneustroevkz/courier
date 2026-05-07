package utils

import (
	"database/sql"
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
