package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string

	MaxConn         int
	MinConn         int
	MaxConnLifeTime time.Duration
	MaxConnIdleTime time.Duration
}

func NewConfig(
	host string, port int,
	user, password, name, ssl string,
	maxConn, minConn int,
	maxConnLifeTime, maxConnIdleTime time.Duration,
) *Config {
	return &Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		Name:            name,
		SSLMode:         ssl,
		MaxConn:         maxConn,
		MinConn:         minConn,
		MaxConnLifeTime: maxConnLifeTime,
		MaxConnIdleTime: maxConnIdleTime,
	}
}

// DSN Returns dsn for postgres clients such as https://github.com/jackc/pgx
func (pc *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pc.Host, pc.Port, pc.User, pc.Password, pc.Name, pc.SSLMode,
	)
}

// MigrationDSN Returns dsn for migration tools such as https://github.com/golang-migrate/migrate
func (pc *Config) MigrationDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pc.User, pc.Password, pc.Host, pc.Port, pc.Name, pc.SSLMode,
	)
}

type Client struct {
	Pool *pgxpool.Pool
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config is not specified")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConn)
	poolConfig.MinConns = int32(cfg.MinConn)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifeTime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{Pool: pool}, nil
}

func (c *Client) Close() {
	c.Pool.Close()
}
