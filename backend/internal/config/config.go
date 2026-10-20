package config

import (
	"os"
	"strconv"
	"time"

	pkgpostgres "backend/pkg/postgres"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	Port             string
	TelegramBotToken string
	LLMBaseURL       string
	LLMAPIKey        string
	LLMModel         string
}

func Load() *Config {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnvWithFallback("DB_PASSWORD", "DB_PASS", "postgres")
	dbName := getEnv("DB_NAME", "expenses")
	dbSSLMode := getEnvWithFallback("DB_SSLMODE", "DB_SSL_MODE", "disable")

	port := getEnv("PORT", "8080")
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	llmBaseURL := getEnv("LLM_BASE_URL", "https://api.openai.com/v1")
	llmAPIKey := os.Getenv("LLM_API_KEY")
	llmModel := getEnv("LLM_MODEL", "gpt-4o-mini")

	return &Config{
		DBHost:           dbHost,
		DBPort:           dbPort,
		DBUser:           dbUser,
		DBPassword:       dbPassword,
		DBName:           dbName,
		DBSSLMode:        dbSSLMode,
		Port:             port,
		TelegramBotToken: tgToken,
		LLMBaseURL:       llmBaseURL,
		LLMAPIKey:        llmAPIKey,
		LLMModel:         llmModel,
	}
}

func (c *Config) PostgresConfig() *pkgpostgres.Config {
	port, err := strconv.Atoi(c.DBPort)
	if err != nil || port <= 0 {
		port = 5432
	}
	return pkgpostgres.NewConfig(
		c.DBHost,
		port,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
		20,
		5,
		time.Hour,
		30*time.Minute,
	)
}

func (c *Config) DSN() string {
	return c.PostgresConfig().DSN()
}

func (c *Config) MigrationDSN() string {
	return c.PostgresConfig().MigrationDSN()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvWithFallback(primaryKey, fallbackKey, defaultVal string) string {
	if val := os.Getenv(primaryKey); val != "" {
		return val
	}
	if val := os.Getenv(fallbackKey); val != "" {
		return val
	}
	return defaultVal
}

