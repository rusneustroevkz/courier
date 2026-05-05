package telegram

import (
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"time"

	"gopkg.in/telebot.v4"
)

type Config struct {
	Token   string `yaml:"token"`
	Timeout int64  `yaml:"timeout"`
}

type Telegram struct {
	bot             *telebot.Bot
	usersRepository users.Querier
}

func NewTelegram(cfg Config, usersRepository users.Querier) (*Telegram, error) {
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
		bot:             bot,
		usersRepository: usersRepository,
	}

	bot.Handle(CommandStart, t.CommandStart)

	logger.Info("telegram bot started", "name", bot.Me.Username)

	return t, nil
}

func (t *Telegram) Start() {
	t.bot.Start()
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
