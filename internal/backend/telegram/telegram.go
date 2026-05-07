package telegram

import (
	"time"

	"github.com/rusneustroevkz/courier/internal/backend/users"
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
		Token:  cfg.Token,
		Poller: &telebot.LongPoller{Timeout: time.Duration(cfg.Timeout) * time.Second},
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

	logger.Info("telegram bot started", "name", bot.Me.Username)

	return t, nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
