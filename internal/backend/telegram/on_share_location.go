package telegram

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/redis"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) OnLocation(ct telebot.Context) error {
	ctx := context.Background()
	log := slog.With("method", "OnLocation")

	msg := ct.Message()
	if msg.Location == nil {
		return nil
	}

	if msg.Location.LivePeriod > 0 {
		expiresTime := time.Now().Add(time.Duration(msg.Location.LivePeriod) * time.Second)
		params := users.SetShareLocation{
			TgUserID:        ct.Sender().ID,
			IsShareLocation: true,
			LivePeriod:      expiresTime,
		}
		if err := t.usersService.SetShareLocation(ctx, params); err != nil {
			log.ErrorContext(ctx, "failed set share location", "err", err)
			return t.SendWithProfile(ct, "Ошибка геопозиции", t.Menu(ct))
		}

		return t.SendWithProfile(ct, "Отлично! Трансляция геопозиции запущена. Теперь вы можете начать смену.", t.Menu(ct))
	}

	return t.SendWithProfile(ct, "Вы отправили статичную точку. Пожалуйста, отправьте именно трансляцию (Живую геопозицию).", t.Menu(ct))
}

func (t *Telegram) OnEditedLocation(ct telebot.Context) error {
	log := slog.With("method", "OnEditedLocation")
	ctx := context.Background()
	msg := ct.Message()

	if msg == nil || msg.Location == nil {
		return t.SendWithProfile(ct, "Ошибка не передалась локация")
	}

	if msg.Location.LivePeriod <= 0 {
		params := users.SetShareLocation{
			TgUserID:        ct.Sender().ID,
			IsShareLocation: false,
			LivePeriod:      time.Now().Add(-1),
		}
		if err := t.usersService.SetShareLocation(ctx, params); err != nil {
			log.ErrorContext(ctx, "failed set share location", "err", err)
			return t.SendWithProfile(ct, "Ошибка геопозиции", t.Menu(ct))
		}

		return t.SendWithProfile(ct, "Трансляция геопозиции остановлена. Смена завершена.", t.Menu(ct))
	}

	userLocation := redis.UserLocation{
		Latitude:  msg.Location.Lat,
		Longitude: msg.Location.Lng,
	}

	jsonData, err := json.Marshal(userLocation)
	if err != nil {
		log.ErrorContext(ctx, "failed to marshal location to json", "err", err)
		return ct.Respond()
	}

	userTgIDString := strconv.Itoa(int(ct.Sender().ID))
	err = t.redisClient.Client.Set(ctx, "location_"+userTgIDString, jsonData, redis.UserLocationTTL).Err()
	if err != nil {
		log.ErrorContext(ctx, "failed set share location", "err", err)
	}

	return ct.Respond()
}
