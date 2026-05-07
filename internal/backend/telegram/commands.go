package telegram

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"gopkg.in/telebot.v4"
	"strconv"
	"strings"
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

	what := strings.Builder{}
	what.WriteString("<b>Профиль</b>")
	what.WriteString("<blockquote>")
	what.WriteString("ID: " + id)

	if user.FullName.Valid {
		what.WriteString("\nИмя: " + user.FullName.String)
	}
	if user.Phone.Valid {
		what.WriteString("\nНомер телефона: " + user.Phone.String)
	}
	what.WriteString("\nВерифицирован: ")
	if user.Verified {
		what.WriteString("да")
	} else {
		what.WriteString("нет")
	}
	if user.Rating.Valid {
		what.WriteString("\nРейтинг: " + user.Rating.String)
	}
	what.WriteString("</blockquote>")

	return ct.Send(what.String(), t.Menu(ct))
}
