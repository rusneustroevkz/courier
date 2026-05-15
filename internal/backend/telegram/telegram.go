package telegram

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/rusneustroevkz/courier/internal/backend/users"
	"gopkg.in/telebot.v4"
)

type Config struct {
	Token   string `yaml:"token"`
	Timeout int64  `yaml:"timeout"`
}

type Telegram struct {
	bot          *telebot.Bot
	usersService users.Service
}

func NewTelegram(cfg Config, usersService users.Service) (*Telegram, error) {
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
		bot:          bot,
		usersService: usersService,
	}

	bot.Handle(CommandStart, t.CommandStart)
	bot.Handle(telebot.OnContact, t.OnContact)
	bot.Handle(telebot.OnCallback, t.OnCallback)

	slog.Info("telegram bot started", "name", bot.Me.Username)

	return t, nil
}

func (t *Telegram) OnCallback(ct telebot.Context) error {
	_ = ct.Notify(telebot.Typing)
	ctx := context.Background()

	parts := strings.Split(ct.Callback().Data, "\f")

	if len(parts) > 1 && strings.HasPrefix(parts[1], CallbackTypeShareContact) {
		return t.CallbackShareContact(parts, ctx, ct)
	}

	return nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
