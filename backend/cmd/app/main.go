package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "backend/internal/adapter/in/http"
	tgadapter "backend/internal/adapter/in/telegram"
	"backend/internal/adapter/out/llm"
	"backend/internal/adapter/out/memory"
	postgresadapter "backend/internal/adapter/out/postgres"
	"backend/internal/app/usecases"
	"backend/internal/config"
	pkgpostgres "backend/pkg/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	slog.Info("configuration loaded", "port", cfg.Port, "db_host", cfg.DBHost, "db_port", cfg.DBPort, "db_name", cfg.DBName, "llm_model", cfg.LLMModel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgCfg := cfg.PostgresConfig()
	runMigrations(pgCfg.MigrationDSN())

	pgClient, err := pkgpostgres.NewClient(ctx, pgCfg)
	if err != nil {
		slog.Error("failed to create postgres client", "error", err)
		os.Exit(1)
	}
	defer pgClient.Close()
	slog.Info("connected to postgresql successfully")

	repo := postgresadapter.NewExpensesRepository(pgClient)
	createExpenseUC := usecases.NewCreateExpenseUC(repo)
	getExpenseUC := usecases.NewGetExpenseUC(repo)
	updateExpenseUC := usecases.NewUpdateExpenseUC(repo)
	deleteExpenseUC := usecases.NewDeleteExpenseUC(repo)
	listExpensesUC := usecases.NewListExpensesUC(repo)
	getDistinctUsersUC := usecases.NewGetDistinctUsersUC(repo)

	memoryStore := memory.NewStore()

	var handleMessageUC *usecases.HandleAgentMessageUC
	if cfg.LLMAPIKey != "" {
		llmClient, err := llm.NewClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
		if err != nil {
			slog.Warn("failed to initialize llm client", "error", err)
		} else {
			toolExecutor := usecases.NewAgentToolExecutor(createExpenseUC, listExpensesUC, updateExpenseUC, deleteExpenseUC)
			handleMessageUC = usecases.NewHandleAgentMessageUC(llmClient, memoryStore, toolExecutor)
			slog.Info("LLM client initialized", "model", cfg.LLMModel, "base_url", cfg.LLMBaseURL)
		}
	} else {
		slog.Warn("LLM_API_KEY is not set, AI features will be disabled")
	}

	httpHandler := httpadapter.NewHandler(
		createExpenseUC,
		getExpenseUC,
		updateExpenseUC,
		deleteExpenseUC,
		listExpensesUC,
		getDistinctUsersUC,
	)
	echoRouter := httpadapter.NewRouter(httpHandler)

	var tgBot *tgadapter.Bot
	if cfg.TelegramBotToken != "" && handleMessageUC != nil {
		bot, err := tgadapter.NewBot(cfg.TelegramBotToken, handleMessageUC)
		if err != nil {
			slog.Warn("failed to initialize telegram bot", "error", err)
		} else {
			tgBot = bot
			slog.Info("telegram bot initialized successfully")
		}
	} else if cfg.TelegramBotToken == "" {
		slog.Warn("TELEGRAM_BOT_TOKEN is not set, telegram bot will not run")
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		addr := ":" + cfg.Port
		slog.Info("starting HTTP server", "address", addr)
		if err := echoRouter.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed", "error", err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		slog.Info("shutting down HTTP server")
		return echoRouter.Shutdown(shutdownCtx)
	})

	if tgBot != nil {
		g.Go(func() error {
			slog.Info("starting Telegram bot long polling")
			tgBot.Start()
			return nil
		})

		g.Go(func() error {
			<-gCtx.Done()
			slog.Info("stopping Telegram bot")
			tgBot.Stop()
			return nil
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("application stopped with error", "error", err)
	} else {
		slog.Info("application gracefully stopped")
	}
}

func runMigrations(dsn string) {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		slog.Warn("migration init warning", "error", err)
		return
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Warn("migration up warning", "error", err)
	} else {
		slog.Info("database migrations applied successfully")
	}
}
