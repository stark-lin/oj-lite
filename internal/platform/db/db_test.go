package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"oj-lite/internal/platform/config"
)

func TestOpenAndMigrate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "oj-lite.db")

	database, err := Open(ctx, config.DBConfig{
		Path:        dbPath,
		BusyTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database file should be created: %v", err)
	}

	if err := Migrate(ctx, database); err != nil {
		t.Fatalf("run initial migration: %v", err)
	}

	if err := Migrate(ctx, database); err != nil {
		t.Fatalf("run idempotent migration: %v", err)
	}

	assertPragmaValue(t, database, "PRAGMA foreign_keys;", 1)
	assertPragmaValue(t, database, "PRAGMA busy_timeout;", 5000)
	assertTextValue(t, database, "PRAGMA journal_mode;", "wal")
	assertCountAtLeast(t, database, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'user_account';", 1)
	assertCountAtLeast(t, database, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'uk_user_account_username';", 1)
	assertCountAtLeast(t, database, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = 'uk_enrollment_student_id';", 1)
	assertCountAtLeast(t, database, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'trigger' AND name = 'trg_submission_consistency_insert';", 1)
}

func TestMigrateRejectsSchemaWithoutSingleClassroomConstraint(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "missing_single_classroom.db")

	database, err := Open(ctx, config.DBConfig{
		Path:        dbPath,
		BusyTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin setup transaction: %v", err)
	}

	if err := execMigration(ctx, tx, "sql/init.sql"); err != nil {
		t.Fatalf("execute init migration: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS uk_user_account_username
		    ON user_account (username);
		CREATE INDEX IF NOT EXISTS idx_classroom_teacher_id
		    ON classroom (teacher_id);
		CREATE INDEX IF NOT EXISTS idx_classroom_current_lesson_id
		    ON classroom (current_lesson_id);
		CREATE UNIQUE INDEX IF NOT EXISTS uk_lesson_sort_order
		    ON lesson (sort_order);
		CREATE UNIQUE INDEX IF NOT EXISTS uk_enrollment_classroom_student
		    ON enrollment (classroom_id, student_id);
		CREATE INDEX IF NOT EXISTS idx_enrollment_student_id
		    ON enrollment (student_id);
		CREATE INDEX IF NOT EXISTS idx_enrollment_current_lesson_id
		    ON enrollment (current_lesson_id);
		CREATE UNIQUE INDEX IF NOT EXISTS uk_lesson_question_lesson_question
		    ON lesson_question (lesson_id, question_id);
		CREATE UNIQUE INDEX IF NOT EXISTS uk_lesson_question_lesson_sort_order
		    ON lesson_question (lesson_id, sort_order);
		CREATE INDEX IF NOT EXISTS idx_lesson_question_question_id
		    ON lesson_question (question_id);
		CREATE INDEX IF NOT EXISTS idx_submission_status_submitted_at
		    ON submission (status, submitted_at DESC);
		CREATE INDEX IF NOT EXISTS idx_submission_enrollment_submitted_at
		    ON submission (enrollment_id, submitted_at DESC);
		CREATE INDEX IF NOT EXISTS idx_submission_lesson_submitted_at
		    ON submission (lesson_id, submitted_at DESC);
		CREATE INDEX IF NOT EXISTS idx_submission_lesson_question_submitted_at
		    ON submission (lesson_question_id, submitted_at DESC);
	`); err != nil {
		t.Fatalf("execute legacy-ready constraints: %v", err)
	}
	if err := execMigration(ctx, tx, "sql/triggers.sql"); err != nil {
		t.Fatalf("execute triggers migration: %v", err)
	}

	teacherResult, err := tx.ExecContext(ctx, `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES ('upgrade_teacher', 'hash', 'teacher', 'active')
	`)
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}
	teacherID, err := teacherResult.LastInsertId()
	if err != nil {
		t.Fatalf("teacher last insert id: %v", err)
	}

	studentResult, err := tx.ExecContext(ctx, `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES ('upgrade_student', 'hash', 'student', 'active')
	`)
	if err != nil {
		t.Fatalf("insert student: %v", err)
	}
	studentID, err := studentResult.LastInsertId()
	if err != nil {
		t.Fatalf("student last insert id: %v", err)
	}

	firstClassroomResult, err := tx.ExecContext(ctx, `
		INSERT INTO classroom (teacher_id, name)
		VALUES (?, 'upgrade_classroom_one')
	`, teacherID)
	if err != nil {
		t.Fatalf("insert first classroom: %v", err)
	}
	firstClassroomID, err := firstClassroomResult.LastInsertId()
	if err != nil {
		t.Fatalf("first classroom last insert id: %v", err)
	}

	secondClassroomResult, err := tx.ExecContext(ctx, `
		INSERT INTO classroom (teacher_id, name)
		VALUES (?, 'upgrade_classroom_two')
	`, teacherID)
	if err != nil {
		t.Fatalf("insert second classroom: %v", err)
	}
	secondClassroomID, err := secondClassroomResult.LastInsertId()
	if err != nil {
		t.Fatalf("second classroom last insert id: %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO enrollment (classroom_id, student_id)
		VALUES (?, ?), (?, ?)
	`, firstClassroomID, studentID, secondClassroomID, studentID); err != nil {
		t.Fatalf("insert duplicate enrollments: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit setup transaction: %v", err)
	}

	assertCountEquals(t, database, `
		SELECT COUNT(*)
		FROM enrollment
		WHERE student_id = ?
	`, 2, studentID)

	err = Migrate(ctx, database)
	if err == nil {
		t.Fatal("migrate should reject schema without single-classroom constraint")
	}
	if got, want := err.Error(), "existing database schema is unsupported"; !strings.Contains(got, want) {
		t.Fatalf("migrate error = %q, want substring %q", got, want)
	}
}

func TestMigrateRejectsLegacySchema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "data", "legacy.db")

	database, err := Open(ctx, config.DBConfig{
		Path:        dbPath,
		BusyTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if _, err := database.ExecContext(ctx, `
		CREATE TABLE user_account (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE classroom (
			id INTEGER PRIMARY KEY,
			teacher_id INTEGER NOT NULL,
			name TEXT NOT NULL
		);
	`); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}

	err = Migrate(ctx, database)
	if err == nil {
		t.Fatal("migrate should reject legacy schema")
	}
	if got, want := err.Error(), "existing database schema is unsupported"; !strings.Contains(got, want) {
		t.Fatalf("legacy migrate error = %q, want substring %q", got, want)
	}
}

func assertPragmaValue(t *testing.T, database *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := database.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}

	if got != want {
		t.Fatalf("query %q = %d, want %d", query, got, want)
	}
}

func assertTextValue(t *testing.T, database *sql.DB, query, want string) {
	t.Helper()

	var got string
	if err := database.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}

	if got != want {
		t.Fatalf("query %q = %q, want %q", query, got, want)
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

func assertCountEquals(t *testing.T, database *sql.DB, query string, want int, args ...any) {
	t.Helper()

	var got int
	if err := database.QueryRowContext(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}

	if got != want {
		t.Fatalf("query %q = %d, want %d", query, got, want)
	}
}
