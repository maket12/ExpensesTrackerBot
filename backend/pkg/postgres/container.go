package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainer struct {
	Container *container.PostgresContainer
	Config    *Config
}

// StartTestContainer Creates and launches a Postgres container for tests.
// Do not use it as a database because the data will be lost once the program stops.
func StartTestContainer(ctx context.Context) (*TestContainer, error) {
	var (
		user     = "user"
		password = "password"
		dbName   = "testdb"
	)

	pgContainer, err := container.Run(ctx,
		"postgres:15-alpine",
		container.WithUsername(user),
		container.WithPassword(password),
		container.WithDatabase(dbName),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil || host == "" {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	cfg := NewConfig(host, int(port.Num()), user, password, dbName,
		"disable", 10, 5, time.Minute, time.Minute,
	)

	return &TestContainer{
		Container: pgContainer,
		Config:    cfg,
	}, nil
}

// Close Terminates the test Postgres container
func (tc *TestContainer) Close(ctx context.Context) error {
	return tc.Container.Terminate(ctx)
}

// MigrateUp Apply migrations up to specified version.
// It searches for migrations in the specified directory
// (make sure you created embed.go file there)
func (tc *TestContainer) MigrateUp(migrationDir embed.FS, version uint) error {
	sourceDriver, err := iofs.New(migrationDir, ".")
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, tc.Config.MigrationDSN())
	if err != nil {
		return fmt.Errorf("failed to init migration tool: %w", err)
	}

	err = m.Migrate(version)
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	var dirtyErr migrate.ErrDirty
	if errors.As(err, &dirtyErr) {
		_ = m.Force(dirtyErr.Version)
		_ = m.Down()
		err = m.Migrate(version)
		if err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	return nil
}

// MigrateDown Rolls back all migrations.
// It searches for migrations in the specified directory
// (make sure you have created embed.go file there)
func (tc *TestContainer) MigrateDown(migrationDir embed.FS) error {
	sourceDriver, err := iofs.New(migrationDir, ".")
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, tc.Config.MigrationDSN())
	if err != nil {
		return fmt.Errorf("failed to init migration tool: %w", err)
	}

	return m.Down()
}

// TruncateTables deletes all data from the specified tables.
// If no tables are specified, it clears all tables available
// in the migration schema.
func (tc *TestContainer) TruncateTables(ctx context.Context, tables ...string) error {
	pool, err := pgxpool.New(ctx, tc.Config.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to postgreSQL for truncate: %w", err)
	}
	defer pool.Close()

	if len(tables) == 0 {
		query := `
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			  AND table_type = 'BASE TABLE' 
			  AND table_name != 'schema_migrations';`

		rows, queryErr := pool.Query(ctx, query)
		if queryErr != nil {
			return fmt.Errorf("failed to fetch table names for truncate: %w", queryErr)
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if scanErr := rows.Scan(&tableName); scanErr != nil {
				return fmt.Errorf("failed to scan table name: %w", scanErr)
			}
			tables = append(tables, tableName)
		}

		if len(tables) == 0 {
			return nil
		}
	}

	quotedTables := make([]string, len(tables))
	for i, t := range tables {
		quotedTables[i] = fmt.Sprintf(`"%s"`, t)
	}

	truncateQuery := fmt.Sprintf(
		"TRUNCATE TABLE %s RESTART IDENTITY CASCADE;",
		strings.Join(quotedTables, ", "),
	)

	if _, err = pool.Exec(ctx, truncateQuery); err != nil {
		return fmt.Errorf("failed to truncate tables [%s]: %w", strings.Join(tables, ", "), err)
	}

	_, _ = pool.Exec(ctx, "DISCARD PLANS;")

	return nil
}
