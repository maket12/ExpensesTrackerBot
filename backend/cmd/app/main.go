package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/cmd/app/config"
	adapterhttp "backend/internal/adapter/in/http"
	tgadapter "backend/internal/adapter/in/telegram"
	"backend/internal/adapter/out/llm"
	"backend/internal/adapter/out/memory"
	postgresadapter "backend/internal/adapter/out/postgres"
	"backend/internal/app/usecases"
	pkgpostgres "backend/pkg/postgres"

	"golang.org/x/sync/errgroup"
)

const (
	shutdownTimeout = 10 * time.Second
)

func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func newLogger(level string) *slog.Logger {
	logLevel := parseLogLevel(level)
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func newPostgresClient(ctx context.Context, cfg *config.Config) (*pkgpostgres.Client, error) {
	pgConfig := pkgpostgres.NewConfig(
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPassword,
		cfg.DBName, cfg.DbSSLMode, cfg.DbMaxConn,
		cfg.DbMinConn, cfg.DbMaxConnLifeTime, cfg.DbMaxConnIdleTime,
	)

	pgClient, err := pkgpostgres.NewClient(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	return pgClient, nil
}

func closePostgresClient(
	ctx context.Context,
	logger *slog.Logger,
	pgClient *pkgpostgres.Client,
) {
	logger.InfoContext(ctx, "closing postgres connection...")
	pgClient.Close()
}

func newLLMClient(
	ctx context.Context,
	cfg *config.Config,
	createUC *usecases.CreateExpenseUC,
	listUC *usecases.ListExpensesUC,
	updateUC *usecases.UpdateExpenseUC,
	deleteUC *usecases.DeleteExpenseUC,
) (*llm.Client, error) {
	return llm.NewClient(
		cfg.LLMBaseURL,
		cfg.LLMAPIKey,
		cfg.LLMModel,
		createUC,
		listUC,
		updateUC,
		deleteUC,
	)
}

func closeLLMClient(
	ctx context.Context,
	logger *slog.Logger,
	llmClient *llm.Client,
) {
	logger.InfoContext(ctx, "closing llm client...")
	if llmClient != nil {
		if err := llmClient.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close llm client", slog.Any("err", err))
		}
	}
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	pgClient, err := newPostgresClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to init postgres client: %w", err)
	}
	defer closePostgresClient(ctx, logger, pgClient)

	// Adapters
	repo := postgresadapter.NewExpensesRepository(pgClient)
	memoryStore := memory.NewStore()

	// Use-cases
	createExpenseUC := usecases.NewCreateExpenseUC(repo)
	getExpenseUC := usecases.NewGetExpenseUC(repo)
	updateExpenseUC := usecases.NewUpdateExpenseUC(repo)
	deleteExpenseUC := usecases.NewDeleteExpenseUC(repo)
	listExpensesUC := usecases.NewListExpensesUC(repo)
	getDistinctUsersUC := usecases.NewGetDistinctUsersUC(repo)

	// LLM Client
	llmClient, err := newLLMClient(
		ctx,
		cfg,
		createExpenseUC,
		listExpensesUC,
		updateExpenseUC,
		deleteExpenseUC,
	)
	if err != nil {
		return fmt.Errorf("failed to init llm client: %w", err)
	}
	defer closeLLMClient(ctx, logger, llmClient)

	handleMessageUC := usecases.NewHandleAgentMessageUC(llmClient, memoryStore)

	// ExpensesHandler & Router
	httpHandler := adapterhttp.NewExpensesHandler(
		logger,
		createExpenseUC,
		getExpenseUC,
		updateExpenseUC,
		deleteExpenseUC,
		listExpensesUC,
		getDistinctUsersUC,
	)

	router := adapterhttp.NewRouter(httpHandler).InitRoutes()

	// Launch telegram bot & server
	var tgBot *tgadapter.Bot
	if cfg.TelegramBotToken != "" {
		tgBot, err = tgadapter.NewBot(cfg.TelegramBotToken, handleMessageUC)
		if err != nil {
			return fmt.Errorf("failed to init telegram bot: %w", err)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HttpPort),
		Handler: router,

		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.InfoContext(gCtx, "starting server", slog.Int("port", cfg.HttpPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.ErrorContext(gCtx, "server failed", slog.Any("err", err))
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		logger.InfoContext(ctx, "shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "graceful shutdown failed", slog.Any("err", err))
			_ = srv.Close() // fallback
			return err
		}
		return nil
	})

	if tgBot != nil {
		g.Go(func() error {
			logger.InfoContext(gCtx, "starting Telegram bot long polling")
			tgBot.Start()
			return nil
		})

		g.Go(func() error {
			<-gCtx.Done()
			logger.InfoContext(ctx, "stopping Telegram bot")
			tgBot.Stop()
			return nil
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.ErrorContext(ctx, "server stopped with error", slog.Any("err", err))
		return err
	}

	logger.InfoContext(ctx, "server exited properly")
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err = runServer(ctx, cfg, logger); err != nil {
		os.Exit(1)
	}
}
