package telegram

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"gopkg.in/telebot.v4"
)

const (
	CommandStart = "/start"
)

func (t *Telegram) CommandStart(ct telebot.Context) error {
	ctx := context.Background()
	log := logger.With("method", "CommandStart")

	_ = t.bot.Notify(ct.Recipient(), telebot.Typing)

	if ct.Sender() == nil {
		log.ErrorContext(ctx, "sender is nil")
		return ct.Send("Ошибка создания пользователя")
	}

	username := ct.Sender().FirstName + " " + ct.Sender().LastName

	createParams := users.CreateParams{
		TgID: sql.NullInt64{
			Int64: ct.Sender().ID,
			Valid: ct.Sender().ID > 0,
		},
		FullName: sql.NullString{
			String: username,
			Valid:  strings.Trim(username, " ") != "",
		},
		Role: users.RoleTypeCourier,
	}
	id, err := t.usersRepository.Create(ctx, createParams)
	if err != nil {
		log.ErrorContext(ctx, "failed create user", "error", err)
		return ct.Send("Ошибка создания пользователя")
	}

	if err := ct.Send("Добро пожаловать в В2В Курьеры"); err != nil {
		return err
	}

	return ct.Send("Пользователь успешно создан, идентификатор: " + strconv.FormatInt(id, 10))
}
