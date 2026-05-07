package telegram

import (
	"context"
	"gopkg.in/telebot.v4"
	"strings"
)

const (
	CallbackTypeShareContact = "share_contact"
)

func (t *Telegram) CallbackShareContact(parts []string, ctx context.Context, ct telebot.Context) error {
	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		return ct.Send("Ошбика коллбэк с поделиться контактом", t.Menu(ct))
	}

	contactMenu := &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}

	btnContact := contactMenu.Contact("Отправить номер телефона")

	contactMenu.Reply(contactMenu.Row(btnContact))

	return ct.Send("Поделитесь с номером телефона", contactMenu)
}
