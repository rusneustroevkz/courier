package telegram

import (
	"context"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
	"log/slog"
	"time"
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
			UserID:          ct.Sender().ID,
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
			UserID:          ct.Sender().ID,
			IsShareLocation: false,
			LivePeriod:      time.Now().Add(-1),
		}
		if err := t.usersService.SetShareLocation(ctx, params); err != nil {
			log.ErrorContext(ctx, "failed set share location", "err", err)
			return ct.Send("Ошибка геопозиции", t.Menu(ct))
		}

		return ct.Send("Трансляция геопозиции остановлена. Смена завершена.")
	}

	lat := msg.Location.Lat
	lng := msg.Location.Lng
	livePeriod := msg.Location.LivePeriod

	slog.Info("Обновление LIVE геопозиции",
		"user_id", msg.Sender.ID,
		"lat", lat,
		"lng", lng,
		"period", livePeriod,
	)

	// TODO: Сохраняем новые координаты в вашу базу данных
	// ctx := context.Background()
	// t.usersService.UpdateLocation(ctx, msg.Sender.ID, lat, lng)

	return nil
}
