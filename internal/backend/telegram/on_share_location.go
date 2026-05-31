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
			return ct.Send("Ошибка геопозиции", t.Menu(ct))
		}

		return ct.Send("Отлично! Трансляция геопозиции запущена. Теперь вы можете начать смену.")
	}

	return ct.Send("Вы отправили статичную точку. Пожалуйста, отправьте именно трансляцию (Живую геопозицию).")
}

func (t *Telegram) OnEditedLocation(ct telebot.Context) error {
	log := slog.With("method", "OnEditedLocation")
	ctx := context.Background()
	msg := ct.Message()

	if msg == nil || msg.Location == nil {
		return ct.Send("Ошибка не передалась локация")
	}

	if msg.Location.Heading == 0 {
		params := users.SetShareLocation{
			TgUserID:        ct.Sender().ID,
			IsShareLocation: false,
			LivePeriod:      time.Now().Add(-1),
		}
		if err := t.usersService.SetShareLocation(ctx, params); err != nil {
			log.ErrorContext(ctx, "failed set share location", "err", err)
			return ct.Send("Ошибка геопозиции", t.Menu(ct))
		}

		return ct.Send("Трансляция геопозиции остановлена. Смена завершена.")
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

	userIDString := strconv.Itoa(int(ct.Sender().ID))
	err = t.redisClient.Client.Set(ctx, "location_"+userIDString, jsonData, redis.UserLocationTTL).Err()
	if err != nil {
		log.ErrorContext(ctx, "failed set share location", "err", err)
	}

	return ct.Respond()
}
