package config

import (
	"fmt"
	"net/url"
	"os"
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

func (c *Config) DSN() string {
	escapedPassword := url.QueryEscape(c.DBPassword)
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, escapedPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
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
