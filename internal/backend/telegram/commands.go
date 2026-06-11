package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/backend/orders"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/redis"
	"gopkg.in/telebot.v4"
	"log/slog"
	"strconv"
)

const (
	CommandStart = "/start"
)

func (t *Telegram) CommandStart(ct telebot.Context) error {
	defer ct.Delete()

	ctx := context.Background()
	log := slog.With("method", "CommandStart")

	defer func() {
		if r := recover(); r != nil {
			log.ErrorContext(ctx, "panic detected", "err", r)
		}
	}()

	_ = t.bot.Notify(ct.Recipient(), telebot.Typing)

	sender := ct.Sender()

	if sender == nil {
		log.ErrorContext(ctx, "sender is nil")
		return t.SendWithProfile(ct, "Ошибка создания пользователя")
	}

	user, err := t.usersService.GetByTgID(ctx, sender.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			params := users.RegisterByTgID{
				UserID:   sender.ID,
				Username: sender.FirstName + " " + sender.LastName,
			}
			user, err = t.usersService.RegisterByTgID(ctx, params)
			if err != nil {
				log.ErrorContext(ctx, "failed to register user", "error", err)
				return t.SendWithProfile(ct, "Ошибка при создании пользователя")
			}
		} else {
			log.ErrorContext(ctx, "failed to get user", "error", err)
			return t.SendWithProfile(ct, "Ошибка при выборке пользователя")
		}
	}

	var opts []MenuOption
	params := make(profileParams)

	var order *orders.GetByIDResult
	if user.OnWork && user.IsShareLocation {
		order, err = t.ordersService.GetActiveOrder(ctx, sender.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.ErrorContext(ctx, "failed to get active order", "error", err)
			return t.SendWithProfile(ct, "Ошибка выборки активного заказа")
		}

		if order != nil {
			courierIDString := fmt.Sprintf("%d", ct.Sender().ID)

			val, err := t.redisClient.Client.Get(ctx, "location_"+courierIDString).Result()
			if err == nil {
				var courierLocation redis.UserLocation

				if err := json.Unmarshal([]byte(val), &courierLocation); err == nil {
					courierLocation = redis.UserLocation{
						Latitude:  62.030696,
						Longitude: 129.741524,
					}

					var pointA, pointB GeoPoint
					var errLat, errLon error
					if order.Status == orders.OrderStatusAccepted {
						pointA.Lat, errLat = strconv.ParseFloat(order.FromLat, 64)
						pointA.Lon, errLon = strconv.ParseFloat(order.FromLon, 64)
					}
					if order.Status == orders.OrderStatusPickedUp {
						pointA.Lat, errLat = strconv.ParseFloat(order.ToLat, 64)
						pointA.Lon, errLon = strconv.ParseFloat(order.ToLon, 64)
					}

					pointB.Lat = float64(courierLocation.Latitude)
					pointB.Lon = float64(courierLocation.Longitude)

					if errLat != nil || errLon != nil {
						log.ErrorContext(ctx, "failed to get active order", "error", errLat, errLon)
					} else {
						dist := DistanceEarth(pointA, pointB)
						params["dist"] = dist
						opts = append(opts, WithDistance(dist), WithOrderStatus(order.Status))
					}
				}
			}
		}
	}

	hasActiveOrder := order != nil && order.ID > 0
	if hasActiveOrder {
		params["order_id"] = order.ID
		params["from_address"] = order.FromAddress
		params["to_address"] = order.ToAddress
		opts = append(opts, WithActiveOrder(order.ID))
	}

	params["id"] = strconv.FormatInt(user.ID, 10)
	params["fullname"] = user.FullName
	params["phone"] = user.Phone
	params["rating"] = user.Rating
	params["has_active_order"] = hasActiveOrder

	if user.OnWork {
		params["on_work_text"] = "на смене"
	} else {
		params["on_work_text"] = "отдыхает"
	}
	if user.Verified {
		params["verified"] = "да"
	} else {
		params["verified"] = "нет"
	}

	what := t.Profile(params)

	t.profilesCache.Store(ct.Sender().ID, what)

	return t.Send(ct, what, t.Menu(ct, opts...))
}
