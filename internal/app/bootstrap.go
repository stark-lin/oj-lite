// Runs the application startup sequence, including config loading, database connection, migrations, and service wiring.

package app

import (
	"context"
	"fmt"
	"os"

	"oj-lite/internal/platform/config"
	"oj-lite/internal/platform/db"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/session"
	"oj-lite/internal/seed"
)

type BootstrapOptions struct {
	SkipSeed bool
}

func Bootstrap() (*App, error) {
	return BootstrapWithOptions(BootstrapOptions{})
}

func BootstrapWithOptions(options BootstrapOptions) (*App, error) {
	cfg := config.Load()
	log := logger.NewLogger(cfg.App.Name)

	databaseExists, err := databaseFileExists(cfg.DB.Path)
	if err != nil {
		return nil, err
	}

	database, err := db.Open(context.Background(), cfg.DB)
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(context.Background(), database); err != nil {
		_ = database.Close()
		return nil, err
	}

	if options.SkipSeed {
		log.Infof("demo seed skipped")
	} else if databaseExists {
		log.Infof("demo seed skipped for existing database")
	} else {
		if err := seed.SeedDemoAccounts(context.Background(), database); err != nil {
			_ = database.Close()
			return nil, err
		}
	}

	log.Infof("database initialized at %s", cfg.DB.Path)

	apiSession, err := session.NewManager()
	if err != nil {
		_ = database.Close()
		return nil, err
	}

	return NewApp(cfg, log, database, apiSession), nil
}

func databaseFileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("stat database path %q: %w", path, err)
	}

	return true, nil
}
