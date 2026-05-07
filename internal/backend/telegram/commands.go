package telegram

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"gopkg.in/telebot.v4"
	"strconv"
)

const (
	CommandStart = "/start"
)

func (t *Telegram) CommandStart(ct telebot.Context) error {
	ctx := context.Background()
	log := logger.With("method", "CommandStart")

	_ = t.bot.Notify(ct.Recipient(), telebot.Typing)

	sender := ct.Sender()

	if sender == nil {
		log.ErrorContext(ctx, "sender is nil")
		return ct.Send("Ошибка создания пользователя")
	}

	user, err := t.usersService.GetByTgID(ctx, sender.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			params := users.RegisterByTgID{
				UserID:   sender.ID,
				Username: sender.FirstName + " " + sender.LastName,
			}
			if err := t.usersService.RegisterByTgID(ctx, params); err != nil {
				log.ErrorContext(ctx, "failed to register user", "error", err)
				return ct.Send("Ошибка при создании пользователя")
			}
		} else {
			log.ErrorContext(ctx, "failed to get user", "error", err)
			return ct.Send("Ошибка при выборке пользователя")
		}
	}

	id := strconv.FormatInt(user.ID, 10)

	return ct.Send("Профиль: "+id, t.Menu(ct))
}
