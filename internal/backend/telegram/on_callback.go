package telegram

import (
	"context"
	"fmt"
	"github.com/rusneustroevkz/courier/internal/backend/orders"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
	"log/slog"
	"strconv"
	"strings"
)

const (
	CallbackShareContact  = "share_contact"
	CallbackOnWork        = "on_work"
	CallbackShareLocation = "on_location"
	CallbackAcceptOrder   = "accept_order"
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

	if len(parts) < 2 {
		log.ErrorContext(ctx, "invalid parts length", "parts_len", len(parts))
		return ct.Send("Ошибка обработки запроса", t.Menu(ct))
	}

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

	if user == nil {
		log.ErrorContext(ctx, "user not found", "user_id", userID, "err", err)
		return ct.Send("Пользователь не найден, обратитесь в поддержку", t.Menu(ct))
	}

	if !user.IsShareLocation {
		return ct.Send("Для начала смены поделитесь геопозицией", t.Menu(ct))
	}

	targetState := !user.OnWork
	params := users.SetOnWork{
		UserID: userID,
		OnWork: targetState,
	}

	// Формируем понятный текст для UI заранее
	var actionWord, successText string
	if targetState {
		actionWord = "открытия"
		successText = "началась"
	} else {
		actionWord = "закрытия"
		successText = "закончилась"
	}

	err = t.usersService.SetOnWork(ctx, params)
	if err != nil {
		log.ErrorContext(ctx, "failed to set on_work", "user_id", userID, "err", err, "target_state", targetState)
		return ct.Send(fmt.Sprintf("Ошибка %s смены, обратитесь в поддержку", actionWord), t.Menu(ct))
	}

	// Если пользователь закрыл смену, не нужно показывать ему доступные заказы
	if !targetState {
		return ct.Send(fmt.Sprintf("Смена %s", successText), t.Menu(ct))
	}

	// 3. Обработка заказов при открытии смены
	list, err := t.ordersService.GetPendingOrders(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to get pending orders", "user_id", userID, "err", err)
		return ct.Send("Ошибка получения списка заказов, обратитесь в поддержку", t.Menu(ct))
	}

	// Ограничиваем количество выводимых заказов во избежание флуда Telegram API
	maxOrdersToShow := 5
	for i, item := range list {
		if i >= maxOrdersToShow {
			_ = ct.Send("...и другие заказы доступны в меню.")
			break
		}

		// Выносим инициализацию разметки за пределы для чистоты
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true}
		menu.Inline(telebot.Row{
			telebot.Btn{Text: "Принять", Unique: CallbackAcceptOrder, Data: strconv.FormatInt(item.ID, 10)},
		})

		// Использование fmt.Sprintf здесь будет лаконичнее и быстрее, чем strings.Builder для 2 строк
		msgText := fmt.Sprintf("Откуда: %s\nКуда: %s", item.FromAddress, item.ToAddress)

		if err = ct.Send(msgText, menu); err != nil {
			log.ErrorContext(ctx, "failed to send pending order", "user_id", userID, "order_id", item.ID, "err", err)
		}
	}

	return ct.Send(fmt.Sprintf("Смена %s", successText), t.Menu(ct))
}

func (t *Telegram) CallbackAcceptOrder(parts []string, ctx context.Context, ct telebot.Context) error {
	log := slog.With("method", "CallbackAcceptOrder")

	if len(parts) < 3 {
		log.ErrorContext(ctx, "invalid parts length", "parts_len", len(parts))
		return ct.Send("Ошибка обработки запроса", t.Menu(ct))
	}

	payload := strings.Split(parts[1], "|")
	if len(payload) < 1 {
		log.Error("invalid callback parts")
		return ct.Send("Ошбика коллбэк", t.Menu(ct))
	}

	orderID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		log.ErrorContext(ctx, "failed to parse order id", "order_id", parts[2], "err", err)
		return ct.Send("Ошибка обработки айди заказа", t.Menu(ct))
	}

	userID := ct.Sender().ID

	acceptOrderParams := orders.AcceptOrder{
		CourierID: userID,
		Status:    orders.OrderStatusAccepted,
		ID:        orderID,
	}
	if err := t.ordersService.AcceptOrder(ctx, acceptOrderParams); err != nil {
		log.ErrorContext(ctx, "failed accept order", "err", err)
		return ct.Send("Ошибка принятия заказа", t.Menu(ct))
	}

	return ct.Send("Заказ принят", t.Menu(ct))
}
