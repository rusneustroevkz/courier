package telegram

import (
	"context"
	"log/slog"

	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) OnContact(ct telebot.Context) error {
	ctx := context.Background()
	log := slog.With("method", "OnContact")

	contact := ct.Message().Contact
	sender := ct.Sender()

	if contact == nil {
		return ct.Send("Передан невалидный контакт", t.Menu(ct))
	}
	if contact.UserID != sender.ID {
		return ct.Send("Номер телефона не совпадает с вашим.", t.Menu(ct))
	}

	args := users.UpdatePhone{
		UserID: sender.ID,
		Phone:  contact.PhoneNumber,
	}
	if err := t.usersService.UpdatePhone(ctx, args); err != nil {
		log.ErrorContext(ctx, "failed update phone number", err)
		return ct.Send("Ошибка при сохранении номера телефона", t.Menu(ct))
	}

	return ct.Send("Номер телефона успешно сохранен", t.Menu(ct))
}
