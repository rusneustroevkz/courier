package utils

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
	"math"
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

type GeoPoint struct {
	Lat, Lon float64 // Широта и долгота в градусах
}

func DistanceEarth(p1, p2 GeoPoint) float64 {
	const earthRadiusKm = 6371.0

	// Переводим градусы в радианы
	lat1 := p1.Lat * math.Pi / 180
	lon1 := p1.Lon * math.Pi / 180
	lat2 := p2.Lat * math.Pi / 180
	lon2 := p2.Lon * math.Pi / 180

	// Разница координат
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Формула гаверсинусов
	a := math.Pow(math.Sin(dLat/2), 2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c * 1000
}
