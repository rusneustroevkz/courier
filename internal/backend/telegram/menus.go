package telegram

import (
	"context"
	"gopkg.in/telebot.v4"
	"log/slog"
)

func (t *Telegram) Menu(ct telebot.Context) *telebot.ReplyMarkup {
	ctx := context.Background()
	sender := ct.Sender()

	log := slog.With("method", "Menu")

	user, err := t.usersService.GetByTgID(ctx, sender.ID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user by telegram id", "error", err)
		return nil
	}

	var rows []telebot.Row

	if !user.Verified {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Ожидайте верификации", Unique: "verification_status"},
		})
	}
	if !user.Phone.Valid {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Поделитесь с номером телефона", Unique: CallbackTypeShareContact},
		})
	}
	if user.Phone.Valid && user.Verified {
		if user.OnWork {
			rows = append(rows, telebot.Row{
				telebot.Btn{Text: "Закончить смену", Unique: CallbackTypeOnWork},
			})
		} else {
			rows = append(rows, telebot.Row{
				telebot.Btn{Text: "Начать смену", Unique: CallbackTypeOnWork},
			})
		}
	}

	if !user.IsShareLocation {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Как поделиться геопозицией", Unique: CallbackTypeShareLocation},
		})
	}

	rows = append(rows, telebot.Row{
		telebot.Btn{Text: "Помощь", Unique: "help"},
	})

	menu := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
		RemoveKeyboard: true,
	}

	menu.Inline(rows...)

	return menu
}
