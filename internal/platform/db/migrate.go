// Initializes the final schema on an empty SQLite database; incremental migrations for old schemas are not supported.

package db

import (
	"context"
	"database/sql"
	"fmt"
)

var bootstrapMigrations = []string{
	"sql/init.sql",
	"sql/constraints.sql",
	"sql/triggers.sql",
}

func Migrate(ctx context.Context, database *sql.DB) error {
	state, err := detectSchemaState(ctx, database)
	if err != nil {
		return err
	}

	switch state {
	case schemaStateReady:
		return nil
	case schemaStateLegacy:
		return fmt.Errorf("existing database schema is unsupported; remove the database file and restart to create a fresh schema")
	case schemaStateFresh:
	default:
		return fmt.Errorf("unknown database schema state")
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}

	for _, name := range bootstrapMigrations {
		if err := execMigration(ctx, tx, name); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration transaction: %w", err)
	}

	return nil
}

type schemaState string

const (
	schemaStateFresh  schemaState = "fresh"
	schemaStateReady  schemaState = "ready"
	schemaStateLegacy schemaState = "legacy"
)

func detectSchemaState(ctx context.Context, database *sql.DB) (schemaState, error) {
	var tableCount int
	if err := database.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
	`).Scan(&tableCount); err != nil {
		return "", fmt.Errorf("detect database tables: %w", err)
	}

	if tableCount == 0 {
		return schemaStateFresh, nil
	}

	ready, err := schemaLooksReady(ctx, database)
	if err != nil {
		return "", err
	}
	if ready {
		return schemaStateReady, nil
	}

	return schemaStateLegacy, nil
}

func schemaLooksReady(ctx context.Context, database *sql.DB) (bool, error) {
	checks := []struct {
		name  string
		query string
	}{
		{
			name: "classroom.current_lesson_id",
			query: `
				SELECT COUNT(*)
				FROM pragma_table_info('classroom')
				WHERE name = 'current_lesson_id'
			`,
		},
		{
			name: "idx_classroom_current_lesson_id",
			query: `
				SELECT COUNT(*)
				FROM sqlite_master
				WHERE type = 'index' AND name = 'idx_classroom_current_lesson_id'
			`,
		},
		{
			name: "trg_enrollment_current_lesson_insert",
			query: `
				SELECT COUNT(*)
				FROM sqlite_master
				WHERE type = 'trigger' AND name = 'trg_enrollment_current_lesson_insert'
			`,
		},
		{
			name: "lesson.creator_id_removed",
			query: `
				SELECT CASE WHEN NOT EXISTS (
					SELECT 1
					FROM pragma_table_info('lesson')
					WHERE name = 'creator_id'
				) THEN 1 ELSE 0 END
			`,
		},
		{
			name: "question.creator_id_removed",
			query: `
				SELECT CASE WHEN NOT EXISTS (
					SELECT 1
					FROM pragma_table_info('question')
					WHERE name = 'creator_id'
				) THEN 1 ELSE 0 END
			`,
		},
		{
			name: "uk_lesson_sort_order",
			query: `
				SELECT COUNT(*)
				FROM sqlite_master
				WHERE type = 'index' AND name = 'uk_lesson_sort_order'
			`,
		},
		{
			name: "uk_enrollment_student_id",
			query: `
				SELECT COUNT(*)
				FROM sqlite_master
				WHERE type = 'index' AND name = 'uk_enrollment_student_id'
			`,
		},
	}

	for _, check := range checks {
		var count int
		if err := database.QueryRowContext(ctx, check.query).Scan(&count); err != nil {
			return false, fmt.Errorf("check database object %q: %w", check.name, err)
		}
		if count == 0 {
			return false, nil
		}
	}

	return true, nil
}

func execMigration(ctx context.Context, tx *sql.Tx, name string) error {
	content, err := MigrationFS.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read migration %q: %w", name, err)
	}

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("execute migration %q: %w", name, err)
	}

	return nil
}
