package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Postgres
	DbHost     string `env:"DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"DB_PORT" envDefault:"5432"`
	DbUser     string `env:"DB_USER" envDefault:"postgres"`
	DbPassword string `env:"DB_PASSWORD" envDefault:"postgres"`
	DBName     string `env:"DB_NAME" envDefault:"postgres"`
	DbSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`

	DbMaxConn         int           `env:"DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// LLM
	LLMBaseURL string `env:"LLM_BASE_URL,required"`
	LLMAPIKey  string `env:"LLM_API_KEY,required"`
	LLMModel   string `env:"LLM_MODEL,required"`

	// Telegram Bot
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN,required"`

	// Service
	HttpPort    int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	fmt.Printf("Config loaded successfully\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   Postgres Host: %s\n", cfg.DbHost)
	fmt.Printf("   LLM Base URL: %s\n", cfg.LLMBaseURL)
	fmt.Printf("   LLM Model: %s\n", cfg.LLMModel)
	fmt.Printf("   HTTP Port: %d\n", cfg.HttpPort)

	return cfg, nil
}

type TestConfig struct {
	// Postgres
	DbHost     string `env:"TEST_DB_HOST" envDefault:"localhost"`
	DbPort     int    `env:"TEST_DB_PORT" envDefault:"5433"`
	DbUser     string `env:"TEST_DB_USER" envDefault:"test"`
	DbPassword string `env:"TEST_DB_PASSWORD" envDefault:"test123"`
	DBName     string `env:"TEST_DB_NAME" envDefault:"test-db"`
	DbSSLMode  string `env:"TEST_DB_SSL_MODE" envDefault:"disable"`

	DbMaxConn         int           `env:"TEST_DB_MAX_CONNECTIONS" envDefault:"30"`
	DbMinConn         int           `env:"TEST_DB_MIN_CONNECTIONS" envDefault:"10"`
	DbMaxConnLifeTime time.Duration `env:"TEST_DB_MAX_CONNECTION_LIFETIME" envDefault:"10m"`
	DbMaxConnIdleTime time.Duration `env:"TEST_DB_MAX_CONNECTION_IDLETIME" envDefault:"5m"`

	// Background jobs
	WebhookPollInterval time.Duration `env:"TEST_WEBHOOK_POLL_INTERVAL" envDefault:"7s"`
	WebhookMaxAttempts  int           `env:"TEST_WEBHOOK_MAX_ATTEMPTS" envDefault:"5"`

	// Service
	HttpPort    int    `env:"TEST_HTTP_PORT" envDefault:"8080"`
	LogLevel    string `env:"TEST_LOG_LEVEL" envDefault:"INFO"`
	Environment string `env:"TEST_ENVIRONMENT" envDefault:"test"`
}

func LoadTest() (*TestConfig, error) {
	cfg := &TestConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load test config: %v", err)
	}

	fmt.Printf("Config loaded successfully\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("   Postgres Host: %s\n", cfg.DbHost)
	fmt.Printf("   HTTP Port: %d\n", cfg.HttpPort)

	return cfg, nil
}
