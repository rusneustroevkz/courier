package telegram

import (
	"context"

	"github.com/rusneustroevkz/courier/internal/admin-frontend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
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
		Token: cfg.Token,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	t := &Telegram{
		bot:          bot,
		usersService: usersService,
	}

	logger.Info("telegram bot started", "name", bot.Me.Username)

	return t, nil
}

func (t *Telegram) Send(ctx context.Context, userID int64, msg string) error {
	chat := &telebot.Chat{ID: userID}
	_, err := t.bot.Send(chat, msg)
	if err != nil {
		logger.ErrorContext(ctx, "failed to send message", "error", err)
		return err
	}
	return nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
