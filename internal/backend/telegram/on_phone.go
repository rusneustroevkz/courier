package telegram

import (
	"context"

	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"gopkg.in/telebot.v4"
)

func (t *Telegram) OnContact(ct telebot.Context) error {
	ctx := context.Background()
	log := logger.With("method", "OnContact")

	contact := ct.Message().Contact
	sender := ct.Sender()

	if contact == nil {
		return ct.Send("Передан невалидный контакт", contactMenu)
	}
	if contact.UserID != sender.ID {
		return ct.Send("Номер телефона не совпадает с вашим.", contactMenu)
	}

	args := users.UpdatePhone{
		UserID: sender.ID,
		Phone:  contact.PhoneNumber,
	}
	if err := t.usersService.UpdatePhone(ctx, args); err != nil {
		log.ErrorContext(ctx, "failed update phone number", err)
		return ct.Send("Ошибка при сохранении номера телефона", contactMenu)
	}

	if err := ct.Send(ct.Recipient(), "Номер телефона успешно сохранен"); err != nil {
		log.ErrorContext(ctx, "failed send to recipient", err)
	}

	text := ct.Message().ReplyTo.Text
	if text == "Поделитесь с номером телефона" {
		return t.CommandStart(ct)
	}

	return nil
}
