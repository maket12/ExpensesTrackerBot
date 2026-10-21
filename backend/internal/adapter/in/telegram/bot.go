package telegram

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"backend/internal/adapter/in/telegram/dto"
	"backend/internal/adapter/in/telegram/errs"
	"backend/internal/adapter/in/telegram/mapper"
	"backend/internal/app/usecases"

	"gopkg.in/telebot.v3"
)

type Bot struct {
	bot             *telebot.Bot
	handleMessageUC *usecases.HandleAgentMessageUC
}

func NewBot(token string, handleMessageUC *usecases.HandleAgentMessageUC) (*Bot, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errs.ErrEmptyToken
	}

	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	tb := &Bot{
		bot:             b,
		handleMessageUC: handleMessageUC,
	}

	tb.setupHandlers()

	return tb, nil
}

func (b *Bot) setupHandlers() {
	b.bot.Handle("/start", func(c telebot.Context) error {
		slog.Info("Telegram /start received", "user_id", c.Sender().ID, "username", c.Sender().Username)
		greeting := `Привет! Я помогу вести учёт расходов.

Просто пиши свои траты обычным текстом:
• «кофе 350»
• «такси 1200 вчера»
• «сколько потратил на еду в этом месяце»
• «последнее было не 350, а 450»
• «удали такси»`
		return c.Send(greeting)
	})

	b.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := strings.TrimSpace(c.Text())
		if text == "" {
			return nil
		}

		slog.Info("Telegram message received", "user_id", c.Sender().ID, "text", text)
		start := time.Now()
		_ = c.Notify(telebot.Typing)

		req := dto.ProcessMessageRequest{
			UserID: c.Sender().ID,
			Text:   text,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		out, err := b.handleMessageUC.Execute(ctx, mapper.ToAgentInput(req))
		if err != nil {
			slog.ErrorContext(ctx, "failed to handle telegram message", "user_id", req.UserID, "error", err, "duration", time.Since(start))
			return c.Send("Не удалось обработать запрос. Пожалуйста, попробуйте сформулировать иначе.")
		}

		resp := mapper.ToTelegramResponse(out)
		slog.InfoContext(ctx, "Telegram message handled successfully", "user_id", req.UserID, "duration", time.Since(start))
		return c.Send(resp.Text)
	})
}

func (b *Bot) Start() {
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}
