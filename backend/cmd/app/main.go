package main

import (
	"context"
	"errors"
	"log"
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
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgCfg := cfg.PostgresConfig()
	runMigrations(pgCfg.MigrationDSN())

	pgClient, err := pkgpostgres.NewClient(ctx, pgCfg)
	if err != nil {
		log.Fatalf("failed to create postgres client: %v", err)
	}
	defer pgClient.Close()

	repo := postgresadapter.NewRepository(pgClient)
	expenseUseCase := usecases.NewExpenseUseCase(repo)
	memoryStore := memory.NewStore(15)

	var agentUseCase *usecases.AgentUseCase
	if cfg.LLMAPIKey != "" {
		llmClient, err := llm.NewClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
		if err != nil {
			log.Printf("warning: failed to initialize llm client: %v", err)
		} else {
			agentUseCase = usecases.NewAgentUseCase(expenseUseCase, llmClient, memoryStore)
		}
	}

	httpHandler := httpadapter.NewHandler(expenseUseCase)
	echoRouter := httpadapter.NewRouter(httpHandler)

	var tgBot *tgadapter.Bot
	if cfg.TelegramBotToken != "" && agentUseCase != nil {
		bot, err := tgadapter.NewBot(cfg.TelegramBotToken, agentUseCase)
		if err != nil {
			log.Printf("warning: failed to initialize telegram bot: %v", err)
		} else {
			tgBot = bot
		}
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		addr := ":" + cfg.Port
		log.Printf("starting HTTP server on %s", addr)
		if err := echoRouter.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Printf("shutting down HTTP server")
		return echoRouter.Shutdown(shutdownCtx)
	})

	if tgBot != nil {
		g.Go(func() error {
			log.Printf("starting Telegram bot long polling")
			tgBot.Start()
			return nil
		})

		g.Go(func() error {
			<-gCtx.Done()
			log.Printf("stopping Telegram bot")
			tgBot.Stop()
			return nil
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("application stopped with error: %v", err)
	} else {
		log.Printf("application gracefully stopped")
	}
}

func runMigrations(dsn string) {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Printf("migration init warning: %v", err)
		return
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("migration up warning: %v", err)
	} else {
		log.Printf("migrations applied successfully")
	}
}
