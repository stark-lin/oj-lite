// Opens the SQLite connection, applies PRAGMA settings such as `foreign_keys=ON`, `journal_mode=WAL`, and `busy_timeout=5000`, and exports the database handle.

package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"oj-lite/internal/platform/config"
)

const driverName = "sqlite"

func Open(ctx context.Context, cfg config.DBConfig) (*sql.DB, error) {
	if err := ensureParentDir(cfg.Path); err != nil {
		return nil, err
	}

	database, err := sql.Open(driverName, cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	// Keep a single pooled connection so PRAGMA session settings stay effective.
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if err := applyPragmas(ctx, database, cfg); err != nil {
		_ = database.Close()
		return nil, err
	}

	return database, nil
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create database directory %q: %w", dir, err)
	}

	return nil
}

func applyPragmas(ctx context.Context, database *sql.DB, cfg config.DBConfig) error {
	if _, err := database.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	var journalMode string
	if err := database.QueryRowContext(ctx, "PRAGMA journal_mode = WAL;").Scan(&journalMode); err != nil {
		return fmt.Errorf("enable wal mode: %w", err)
	}

	if _, err := database.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d;", cfg.BusyTimeout.Milliseconds())); err != nil {
		return fmt.Errorf("set busy timeout: %w", err)
	}

	return nil
}
