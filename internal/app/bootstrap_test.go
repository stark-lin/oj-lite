package app

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestBootstrapWithOptionsSeedsFreshDatabaseByDefault(t *testing.T) {
	t.Setenv("APP_NAME", "oj-lite-test")
	t.Setenv("APP_ENV", "test")
	t.Setenv("DB_PATH", filepath.Join(t.TempDir(), "oj-lite.db"))
	t.Setenv("GIN_MODE", "test")

	application, err := BootstrapWithOptions(BootstrapOptions{})
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	defer shutdownTestApp(t, application)

	assertCountEquals(t, application.DB(), `
		SELECT COUNT(*)
		FROM user_account
		WHERE username IN ('teacher', 'student');
	`, 2)
	assertCountEquals(t, application.DB(), "SELECT COUNT(*) FROM lesson;", 24)
	assertCountAtLeast(t, application.DB(), "SELECT COUNT(*) FROM lesson_question;", 24)
}

func TestBootstrapWithOptionsSkipsSeed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "oj-lite.db")

	t.Setenv("APP_NAME", "oj-lite-test")
	t.Setenv("APP_ENV", "test")
	t.Setenv("DB_PATH", dbPath)
	t.Setenv("GIN_MODE", "test")

	application, err := BootstrapWithOptions(BootstrapOptions{
		SkipSeed: true,
	})
	if err != nil {
		t.Fatalf("bootstrap with skip seed: %v", err)
	}

	assertCountEquals(t, application.DB(), "SELECT COUNT(*) FROM user_account;", 0)
	assertCountAtLeast(t, application.DB(), "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'user_account';", 1)

	shutdownTestApp(t, application)

	application, err = BootstrapWithOptions(BootstrapOptions{})
	if err != nil {
		t.Fatalf("bootstrap existing database: %v", err)
	}
	defer shutdownTestApp(t, application)

	assertCountEquals(t, application.DB(), "SELECT COUNT(*) FROM user_account;", 0)
}

func assertCountEquals(t *testing.T, database *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := database.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}

	if got != want {
		t.Fatalf("query %q = %d, want %d", query, got, want)
	}
}

func assertCountAtLeast(t *testing.T, database *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := database.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}

	if got < want {
		t.Fatalf("query %q = %d, want at least %d", query, got, want)
	}
}
