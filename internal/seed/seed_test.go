package seed

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"oj-lite/internal/platform/config"
	platformdb "oj-lite/internal/platform/db"
)

func TestSeedDemoAccounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	database := openSeedTestDB(t, ctx)
	defer database.Close()

	if err := SeedDemoAccounts(ctx, database); err != nil {
		t.Fatalf("seed demo accounts: %v", err)
	}

	if err := SeedDemoAccounts(ctx, database); err != nil {
		t.Fatalf("run idempotent demo seed: %v", err)
	}

	assertTextValue(t, database, "SELECT role FROM user_account WHERE username = 'teacher';", "teacher")
	assertTextValue(t, database, "SELECT role FROM user_account WHERE username = 'student';", "student")
	assertCountAtLeast(t, database, "SELECT COUNT(*) FROM classroom WHERE name = 'teacher_demo_classroom';", 1)
	assertCountAtLeast(t, database, "SELECT COUNT(*) FROM classroom WHERE name = 'example_classroom';", 1)
	assertCountAtLeast(t, database, `
		SELECT COUNT(*)
		FROM lesson
		WHERE title = 'example_lesson';
	`, 1)
	assertCountAtLeast(t, database, `
		SELECT COUNT(*)
		FROM question
		WHERE title = 'example_question';
	`, 1)
	assertCountAtLeast(t, database, `
		SELECT COUNT(*)
		FROM lesson_question lq
		JOIN lesson l ON l.id = lq.lesson_id
		JOIN question q ON q.id = lq.question_id
		WHERE l.title = 'example_lesson' AND q.title = 'example_question';
	`, 1)
	assertCountEquals(t, database, `
		SELECT COUNT(*)
		FROM enrollment e
		JOIN user_account s ON s.id = e.student_id
		JOIN classroom c ON c.id = e.classroom_id
		JOIN lesson l ON l.id = e.current_lesson_id
		WHERE s.username = 'student'
		  AND c.name = 'teacher_demo_classroom'
		  AND l.title = 'example_lesson';
	`, 1)
	assertCountEquals(t, database, `
		SELECT COUNT(*)
		FROM enrollment e
		JOIN user_account s ON s.id = e.student_id
		WHERE s.username = 'student';
	`, 1)

	var demoClassroomID, exampleClassroomID, studentID int64
	if err := database.QueryRowContext(ctx, `SELECT id FROM classroom WHERE name = 'teacher_demo_classroom' LIMIT 1`).Scan(&demoClassroomID); err != nil {
		t.Fatalf("load demo classroom id: %v", err)
	}
	if err := database.QueryRowContext(ctx, `SELECT id FROM classroom WHERE name = 'example_classroom' LIMIT 1`).Scan(&exampleClassroomID); err != nil {
		t.Fatalf("load example classroom id: %v", err)
	}
	if err := database.QueryRowContext(ctx, `SELECT id FROM user_account WHERE username = 'student' LIMIT 1`).Scan(&studentID); err != nil {
		t.Fatalf("load demo student id: %v", err)
	}
	if _, err := database.ExecContext(ctx, `
		INSERT INTO enrollment (classroom_id, student_id)
		VALUES (?, ?)
	`, exampleClassroomID, studentID); err == nil {
		t.Fatal("second enrollment for the same student should fail")
	}

	var actualClassroomID int64
	if err := database.QueryRowContext(ctx, `
		SELECT classroom_id
		FROM enrollment
		WHERE student_id = ?
		LIMIT 1
	`, studentID).Scan(&actualClassroomID); err != nil {
		t.Fatalf("load enrolled classroom id: %v", err)
	}
	if actualClassroomID != demoClassroomID {
		t.Fatalf("enrolled classroom id = %d, want %d", actualClassroomID, demoClassroomID)
	}
}

func openSeedTestDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	database, err := platformdb.Open(ctx, config.DBConfig{
		Path:        filepath.Join(t.TempDir(), "data", "seed.db"),
		BusyTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	if err := platformdb.Migrate(ctx, database); err != nil {
		_ = database.Close()
		t.Fatalf("migrate database: %v", err)
	}

	return database
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
