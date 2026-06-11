package telegram

import (
	"context"
	"github.com/rusneustroevkz/courier/internal/backend/orders"
	"gopkg.in/telebot.v4"
	"log/slog"
	"strconv"
)

type MenuConfig struct {
	activeOrderID int64
	orderStatus   orders.OrderStatus
	dist          float64
}

type MenuOption func(*MenuConfig)

func WithActiveOrder(orderID int64) MenuOption {
	return func(c *MenuConfig) {
		c.activeOrderID = orderID
	}
}

func WithDistance(dist float64) MenuOption {
	return func(c *MenuConfig) {
		c.dist = dist
	}
}

func WithOrderStatus(status orders.OrderStatus) MenuOption {
	return func(c *MenuConfig) {
		c.orderStatus = status
	}
}

func (t *Telegram) Menu(ct telebot.Context, opts ...MenuOption) *telebot.ReplyMarkup {
	ctx := context.Background()
	sender := ct.Sender()

	config := &MenuConfig{}

	for _, opt := range opts {
		opt(config)
	}

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
	if user.Phone == "" {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Поделитесь с номером телефона", Unique: CallbackShareContact},
		})
	}
	if config.activeOrderID == 0 && user.Phone != "" && user.Verified {
		if user.OnWork {
			rows = append(rows,
				telebot.Row{telebot.Btn{Text: "Закончить смену", Unique: CallbackOnWork}},
				telebot.Row{telebot.Btn{Text: "Список заказов", Unique: CallbackOrdersList}},
			)
		} else {
			rows = append(rows, telebot.Row{
				telebot.Btn{Text: "Начать смену", Unique: CallbackOnWork},
			})
		}
	}
	if config.activeOrderID > 0 && config.orderStatus == orders.OrderStatusPickedUp && config.dist < 100 {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Завершить заказ", Unique: CallbackDoneOrder, Data: strconv.FormatInt(config.activeOrderID, 10)},
		})
	}
	if config.activeOrderID > 0 && config.orderStatus == orders.OrderStatusAccepted && config.dist < 100 {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Забрать заказ", Unique: CallbackPickUpOrder, Data: strconv.FormatInt(config.activeOrderID, 10)},
		})
	}

	if !user.IsShareLocation {
		rows = append(rows, telebot.Row{
			telebot.Btn{Text: "Как поделиться геопозицией", Unique: CallbackShareLocation},
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
