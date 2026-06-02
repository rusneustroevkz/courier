package telegram

import (
	"context"
	"log/slog"
	"strings"
	"sync"
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
	profilesCache sync.Map
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
		profilesCache: sync.Map{},
	}

	bot.Handle(CommandStart, t.CommandStart)
	bot.Handle(telebot.OnContact, t.OnContact)
	bot.Handle(telebot.OnCallback, t.OnCallback)
	bot.Handle(telebot.OnLocation, t.OnLocation)
	bot.Handle(telebot.OnEdited, t.OnEditedLocation)

	slog.Info("telegram bot started", "name", bot.Me.Username)

	return t, nil
}

func (t *Telegram) Send(ct telebot.Context, what string, opts ...interface{}) error {
	err := ct.Edit(what, opts...)
	if err != nil && !errors.Is(err, telebot.ErrBadContext) {
		slog.Error("telegram bot editing error", "error", err)
		return ct.Send("Что-то пошло нет так, попробуйте позже")
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

func (t *Telegram) SendWithProfile(ct telebot.Context, what string, opts ...interface{}) error {
	err := ct.Edit(what, opts...)
	if err != nil && !errors.Is(err, telebot.ErrBadContext) {
		return err
	}

	profile, ok := t.profilesCache.Load(ct.Sender().ID)
	if ok {
		what += "\n\n" + profile.(string)
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
	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackOrdersList) {
		return t.CallbackOrdersList(parts, ctx, ct)
	}

	return nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
