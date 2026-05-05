package telegram

import "gopkg.in/telebot.v4"

const (
	CommandStart = "/start"
)

func (t *Telegram) CommandStart(ct telebot.Context) error {
	_ = t.bot.Notify(ct.Recipient(), telebot.Typing)

	err := ct.Send("Добро пожаловать в В2В Курьеры")
	if err != nil {
		return err
	}

	return nil
}
