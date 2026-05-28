package telegram

import (
	"context"
	"fmt"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
	"log/slog"
	"strings"
)

const (
	CallbackTypeShareContact  = "share_contact"
	CallbackTypeOnWork        = "on_work"
	CallbackTypeShareLocation = "on_location"
)

func (t *Telegram) CallbackShareContact(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackShareContact")
	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return ct.Send("Ошбика коллбэк с поделиться контактом", t.Menu(ct))
	}

	contactMenu := &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}

	btnContact := contactMenu.Contact("Отправить номер телефона")

	contactMenu.Reply(contactMenu.Row(btnContact))

	return ct.Send("Поделитесь с номером телефона", contactMenu)
}

func (t *Telegram) CallbackShareLocation(parts []string, ctx context.Context, ct telebot.Context) error {
	defer ct.Respond()

	instruction := "📍 *Как отправить живую геопозицию:*\n\n" +
		"1. Нажмите на иконку *скрепки* (или `+`) рядом с полем ввода сообщения.\n" +
		"2. Выберите пункт *'Геопозиция'* (Location).\n" +
		"3. Нажмите *'Транслировать мою геопозицию'* (Share My Live Location) и выберите время (например, 8 часов)."

	return ct.Send(instruction, telebot.ModeMarkdown)
}

func (t *Telegram) CallbackOnWork(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackOnWork")
	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return ct.Send("Ошбика коллбэк", t.Menu(ct))
	}

	userID := ct.Sender().ID
	user, err := t.usersService.GetByTgID(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user by telegram id", "user_id", userID, "err", err)
		return ct.Send("Ошибка выборки пользователя, обратитесь в поддержку", t.Menu(ct))
	}

	if user == nil || user.ID == 0 {
		log.ErrorContext(ctx, "user id is nil", "user_id", userID, "err", err)
		return ct.Send("Невалидный пользователь, обратитесь в поддержку", t.Menu(ct))
	}

	if !user.IsShareLocation {
		return ct.Send("Для начала смены поделитесь геопозицией", t.Menu(ct))
	}

	params := users.SetOnWork{
		UserID: userID,
		OnWork: true,
	}
	errorText := ""
	successText := ""
	if user.OnWork {
		params.OnWork = true
		errorText = "начала"
		successText = "началась"
	} else {
		params.OnWork = false
		errorText = "конца"
		successText = "закончилась"
	}
	err = t.usersService.SetOnWork(ctx, params)
	if err != nil {
		log.ErrorContext(ctx, "failed to set on_work", "user_id", userID, "err", err)
		return ct.Send(fmt.Sprintf("Ошибка %s смены, обратитесь в поддержку", errorText), t.Menu(ct))
	}

	list, err := t.ordersService.GetPendingOrders(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to get pending orders", "user_id", userID, "err", err)
		return ct.Send("Ошибка списка заказов, обратитесь в поддержку", t.Menu(ct))
	}

	var rows []telebot.Row

	rows = append(rows, telebot.Row{
		telebot.Btn{Text: "Принять", Unique: "accept_order"},
	})

	menu := &telebot.ReplyMarkup{
		ResizeKeyboard: true,
		RemoveKeyboard: true,
	}

	menu.Inline(rows...)

	for _, item := range list {
		what := strings.Builder{}
		what.WriteString("Откуда: " + item.FromAddress)
		what.WriteString("\nКуда: " + item.ToAddress)

		if err = ct.Send(what.String(), menu); err != nil {
			log.ErrorContext(ctx, "failed to send pending order", "user_id", userID, "err", err)
		}
	}

	return ct.Send(fmt.Sprintf("Смена %s", successText), t.Menu(ct))
}
