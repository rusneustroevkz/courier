package telegram

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
	"log/slog"
	"strconv"
	"strings"
)

const (
	CommandStart = "/start"
)

func (t *Telegram) CommandStart(ct telebot.Context) error {
	defer ct.Delete()

	ctx := context.Background()
	log := slog.With("method", "CommandStart")

	_ = t.bot.Notify(ct.Recipient(), telebot.Typing)

	sender := ct.Sender()

	if sender == nil {
		log.ErrorContext(ctx, "sender is nil")
		return t.Send(ct, "Ошибка создания пользователя")
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
				return t.Send(ct, "Ошибка при создании пользователя")
			}
		} else {
			log.ErrorContext(ctx, "failed to get user", "error", err)
			return t.Send(ct, "Ошибка при выборке пользователя")
		}
	}

	order, err := t.ordersService.GetActiveOrder(ctx, sender.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.ErrorContext(ctx, "failed to get active order", "error", err)
		return t.Send(ct, "Ошибка выборки активного заказа")
	}

	hasActiveOrder := order != nil && order.ID > 0

	id := strconv.FormatInt(user.ID, 10)

	what := strings.Builder{}
	what.WriteString("<b>Профиль</b>")
	what.WriteString("<blockquote>")
	what.WriteString("ID: " + id)

	onWorkText := ""
	if user.OnWork {
		onWorkText = "на смене"
	} else {
		onWorkText = "отдыхает"
	}
	what.WriteString("\nСмена: " + onWorkText)

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

	var activeOrderID int64
	if hasActiveOrder {
		activeOrderID = order.ID

		what.WriteString("\n\n<b>Активный заказ</b>")
		what.WriteString("<blockquote>")
		what.WriteString("От: " + order.FromAddress)
		what.WriteString("\nДо: " + order.ToAddress)
		what.WriteString("</blockquote>")
	}

	return t.Send(ct, what.String(), t.Menu(ct, WithActiveOrder(activeOrderID)))
}
