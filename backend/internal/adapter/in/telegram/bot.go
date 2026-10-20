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
	bot   *telebot.Bot
	agent *usecases.AgentUseCase
}

func NewBot(token string, agent *usecases.AgentUseCase) (*Bot, error) {
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
		bot:   b,
		agent: agent,
	}

	tb.setupHandlers()

	return tb, nil
}

func (b *Bot) setupHandlers() {
	b.bot.Handle("/start", func(c telebot.Context) error {
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

		_ = c.Notify(telebot.Typing)

		req := dto.ProcessMessageRequest{
			UserID: c.Sender().ID,
			Text:   text,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		out, err := b.agent.HandleMessage(ctx, mapper.ToAgentInput(req))
		if err != nil {
			slog.ErrorContext(ctx, "failed to handle message", "error", err)
			return c.Send("Не удалось обработать запрос. Пожалуйста, попробуйте сформулировать иначе.")
		}

		resp := mapper.ToTelegramResponse(out)
		return c.Send(resp.Text)
	})
}

func (b *Bot) Start() {
	b.bot.Start()
}

func (b *Bot) Stop() {
	b.bot.Stop()
}
