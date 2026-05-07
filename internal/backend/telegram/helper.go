package telegram

import "gopkg.in/telebot.v4"

var (
	contactMenu = &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
)

func (t *Telegram) registerPhone(ct telebot.Context) error {
	btnContact := contactMenu.Contact("Отправить номер телефона")

	contactMenu.Reply(contactMenu.Row(btnContact))

	return ct.Send("Поделитесь с номером телефона", contactMenu)
}
