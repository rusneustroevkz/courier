package telegram

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/backend/orders"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/redis"
	"gopkg.in/telebot.v4"
)

type Config struct {
	Token   string `yaml:"token"`
	Timeout int64  `yaml:"timeout"`
}

type Telegram struct {
	bot           *telebot.Bot
	usersService  users.Service
	ordersService orders.Service
	redisClient   *redis.Redis
}

func NewTelegram(cfg Config, usersService users.Service, ordersService orders.Service, redisClient *redis.Redis) (*Telegram, error) {
	pref := telebot.Settings{
		Token:     cfg.Token,
		Poller:    &telebot.LongPoller{Timeout: time.Duration(cfg.Timeout) * time.Second},
		ParseMode: telebot.ModeHTML,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	commands := []telebot.Command{
		{Text: CommandStart, Description: "Перезапустить бота"},
	}

	if err := bot.SetCommands(commands); err != nil {
		return nil, err
	}

	t := &Telegram{
		bot:           bot,
		usersService:  usersService,
		ordersService: ordersService,
		redisClient:   redisClient,
	}

	bot.Handle(CommandStart, t.CommandStart)
	bot.Handle(telebot.OnContact, t.OnContact)
	bot.Handle(telebot.OnCallback, t.OnCallback)
	bot.Handle(telebot.OnLocation, t.OnLocation)
	bot.Handle(telebot.OnEdited, t.OnEditedLocation)

	slog.Info("telegram bot started", "name", bot.Me.Username)

	return t, nil
}

func (t *Telegram) Send(ct telebot.Context, what interface{}, opts ...interface{}) error {
	err := ct.Edit(what, opts...)
	if err != nil && !errors.Is(err, telebot.ErrBadContext) {
		return err
	}

	msg, err := t.bot.Send(ct.Recipient(), what, opts...)
	if err != nil {
		slog.Error("error send message", "err", err.Error())
		return ct.Send("Что-то пошло нет так, попробуйте позже")
	}

	time.AfterFunc(time.Minute*2, func() {
		err = t.bot.Delete(msg)
		if err != nil {
			slog.Error("error delete message", "err", err.Error())
		}
	})

	return nil
}

func (t *Telegram) OnCallback(ct telebot.Context) error {
	_ = ct.Notify(telebot.Typing)
	ctx := context.Background()

	parts := strings.Split(ct.Callback().Data, "\f")

	if ct.Message().Location == nil {
		params := users.SetShareLocation{
			TgUserID:        ct.Sender().ID,
			IsShareLocation: false,
			LivePeriod:      time.Now().Add(-1),
			OnWork:          false,
		}
		if err := t.usersService.SetShareLocation(ctx, params); err != nil {
			slog.ErrorContext(ctx, "failed to set active order", "error", err)
		}
	}

	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackShareContact) {
		return t.CallbackShareContact(parts, ctx, ct)
	}
	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackOnWork) {
		return t.CallbackOnWork(parts, ctx, ct)
	}
	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackShareLocation) {
		return t.CallbackShareLocation(parts, ctx, ct)
	}
	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackAcceptOrder) {
		return t.CallbackAcceptOrder(parts, ctx, ct)
	}
	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackDoneOrder) {
		return t.CallbackDoneOrder(parts, ctx, ct)
	}

	return nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
