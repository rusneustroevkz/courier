package telegram

import (
	"context"
	"log/slog"

	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) OnContact(ct telebot.Context) error {
	_ = t.bot.Notify(ct.Recipient(), telebot.Typing)

	ctx := context.Background()
	log := slog.With("method", "OnContact")

	contact := ct.Message().Contact
	sender := ct.Sender()

	if contact == nil {
		return t.Send(ct, "Передан невалидный контакт", t.Menu(ct))
	}
	if contact.UserID != sender.ID {
		return t.Send(ct, "Номер телефона не совпадает с вашим.", t.Menu(ct))
	}

	args := users.UpdatePhone{
		UserID: sender.ID,
		Phone:  contact.PhoneNumber,
	}
	if err := t.usersService.UpdatePhone(ctx, args); err != nil {
		log.ErrorContext(ctx, "failed update phone number", err)
		return t.Send(ct, "Ошибка при сохранении номера телефона", t.Menu(ct))
	}

	return t.Send(ct, "Номер телефона успешно сохранен", t.Menu(ct))
}
