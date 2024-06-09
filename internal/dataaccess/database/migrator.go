package database

import (
	"context"
	"database/sql"
	"embed"

	"github.com/nhtuan0700/GoLoad/internal/configs"
	"github.com/nhtuan0700/GoLoad/internal/utils"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

var (
	//go:embed migrations/mysql/*
	migrationDirectorMYSQL embed.FS
)

type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}

type migrator struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewMigrator(
	db *sql.DB,
	logger *zap.Logger,
) Migrator {
	return &migrator{
		db:     db,
		logger: logger,
	}
}

func (m *migrator) migrate(ctx context.Context, direction migrate.MigrationDirection) error {
	logger := utils.LoggerWithContext(ctx, m.logger).With(zap.Int("direction", int(direction)))

	migrationCount, err := migrate.ExecContext(ctx, m.db, string(configs.DatabaseTypeMySQL), migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationDirectorMYSQL,
		Root:       "migrations/mysql",
	}, direction)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to executed migrations")
		return err
	}

	logger.With(zap.Int("migration_count", migrationCount)).Info("successfully executed database migration")
	return nil
}

func (m *migrator) Up(ctx context.Context) error {
	return m.migrate(ctx, migrate.Up)
}

func (m *migrator) Down(ctx context.Context) error {
	return m.migrate(ctx, migrate.Down)
}
