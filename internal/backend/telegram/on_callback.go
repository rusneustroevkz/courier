package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/rusneustroevkz/courier/internal/backend/orders"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
)

const (
	CallbackShareContact  = "share_contact"
	CallbackOnWork        = "on_work"
	CallbackShareLocation = "on_location"
	CallbackAcceptOrder   = "accept_order"
	CallbackDoneOrder     = "done_order"
)

func (t *Telegram) CallbackShareContact(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackShareContact")
	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return t.Send(ct, "Ошбика коллбэк с поделиться контактом", t.Menu(ct))
	}

	contactMenu := &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}

	btnContact := contactMenu.Contact("Отправить номер телефона")

	contactMenu.Reply(contactMenu.Row(btnContact))

	if err := ct.Respond(); err != nil {
		log.Error("failed to respond to callback", "error", err)
	}

	_, err := t.bot.Send(ct.Recipient(), "Пожалуйста, поделитесь номером телефона, нажав на кнопку ниже:", contactMenu)
	return err
}

func (t *Telegram) CallbackShareLocation(parts []string, ctx context.Context, ct telebot.Context) error {
	instruction := "📍 *Как отправить живую геопозицию:*\n\n" +
		"1. Нажмите на иконку *скрепки* (или `+`) рядом с полем ввода сообщения.\n" +
		"2. Выберите пункт *'Геопозиция'* (Location).\n" +
		"3. Нажмите *'Транслировать мою геопозицию'* (Share My Live Location) и выберите время (например, 8 часов)."

	return t.Send(ct, instruction, telebot.ModeMarkdown)
}

func (t *Telegram) CallbackOnWork(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackOnWork")

	if len(parts) < 2 {
		log.ErrorContext(ctx, "invalid parts length", "parts_len", len(parts))
		return t.Send(ct, "Ошибка обработки запроса", t.Menu(ct))
	}

	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return t.Send(ct, "Ошбика коллбэк", t.Menu(ct))
	}

	userID := ct.Sender().ID
	user, err := t.usersService.GetByTgID(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user by telegram id", "user_id", userID, "err", err)
		return t.Send(ct, "Ошибка выборки пользователя, обратитесь в поддержку", t.Menu(ct))
	}

	if user == nil {
		log.ErrorContext(ctx, "user not found", "user_id", userID, "err", err)
		return t.Send(ct, "Пользователь не найден, обратитесь в поддержку", t.Menu(ct))
	}

	if !user.IsShareLocation {
		return ct.Edit("Для начала смены поделитесь геопозицией", t.Menu(ct))
	}

	targetState := !user.OnWork
	params := users.SetOnWork{
		UserID: userID,
		OnWork: targetState,
	}

	err = t.usersService.SetOnWork(ctx, params)
	if err != nil {
		log.ErrorContext(ctx, "failed to set on_work", "user_id", userID, "err", err, "target_state", targetState)
		return t.Send(ct, "Ошибка смены, обратитесь в поддержку", t.Menu(ct))
	}

	if !targetState {
		return t.CommandStart(ct)
	}

	list, err := t.ordersService.GetPendingOrders(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to get pending orders", "user_id", userID, "err", err)
		return t.Send(ct, "Ошибка получения списка заказов, обратитесь в поддержку", t.Menu(ct))
	}

	maxOrdersToShow := 5
	for i, item := range list {
		if i >= maxOrdersToShow {
			_ = t.Send(ct, "...и другие заказы доступны в меню.")
			break
		}

		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		menu.Inline(telebot.Row{
			telebot.Btn{Text: "Принять", Unique: CallbackAcceptOrder, Data: strconv.FormatInt(item.ID, 10)},
		})

		msgText := fmt.Sprintf("Откуда: %s\nКуда: %s", item.FromAddress, item.ToAddress)

		if err = t.Send(ct, msgText, menu); err != nil {
			log.ErrorContext(ctx, "failed to send pending order", "user_id", userID, "order_id", item.ID, "err", err)
		}
	}

	return t.CommandStart(ct)
}

func (t *Telegram) CallbackAcceptOrder(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackAcceptOrder")

	if len(parts) < 2 {
		log.ErrorContext(ctx, "invalid parts length", "parts_len", len(parts))
		return t.Send(ct, "Ошибка обработки запроса", t.Menu(ct))
	}

	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return t.Send(ct, "Ошбика коллбэк", t.Menu(ct))
	}

	orderID, err := strconv.ParseInt(payload[1], 10, 64)
	if err != nil {
		log.ErrorContext(ctx, "failed to parse order id", "order_id", parts[2], "err", err)
		return t.Send(ct, "Ошибка обработки айди заказа", t.Menu(ct))
	}

	userID := ct.Sender().ID

	acceptOrderParams := orders.AcceptOrder{
		CourierID: userID,
		Status:    orders.OrderStatusAccepted,
		ID:        orderID,
	}
	if err := t.ordersService.AcceptOrder(ctx, acceptOrderParams); err != nil {
		log.ErrorContext(ctx, "failed accept order", "err", err)
		return t.Send(ct, "Ошибка принятия заказа", t.Menu(ct))
	}

	return t.CommandStart(ct)
}

func (t *Telegram) CallbackDoneOrder(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackDoneOrder")

	if len(parts) < 2 {
		log.ErrorContext(ctx, "invalid parts length", "parts_len", len(parts))
		return t.Send(ct, "Ошибка обработки запроса", t.Menu(ct))
	}

	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return t.Send(ct, "Ошбика коллбэк", t.Menu(ct))
	}

	orderID, err := strconv.ParseInt(payload[1], 10, 64)
	if err != nil {
		log.ErrorContext(ctx, "failed to parse order id", "order_id", parts[2], "err", err)
		return t.Send(ct, "Ошибка обработки айди заказа", t.Menu(ct))
	}

	err = t.ordersService.DoneOrder(ctx, orderID)
	if err != nil {
		log.ErrorContext(ctx, "failed done order", "order_id", orderID, "err", err)
		return t.Send(ct, "Ошибка завершении заказа", t.Menu(ct))
	}

	return t.CommandStart(ct)
}
