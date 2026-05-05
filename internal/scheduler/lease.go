// Claims tasks and advances states to avoid processing the same submission twice.

package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"oj-lite/internal/platform/logger"
)

const staleJudgingTimeout = time.Minute

type claimedSubmission struct {
	ID            int64
	SourceCode    string
	ReferenceCode string
	TestCases     string
}

type lease struct {
	db  *sql.DB
	log *logger.Logger
}

func newLease(database *sql.DB, log *logger.Logger) *lease {
	return &lease{db: database, log: log}
}

func (lease *lease) claimPendingSubmissions(ctx context.Context, limit int) ([]claimedSubmission, error) {
	if limit <= 0 {
		return nil, nil
	}

	if err := lease.expireStaleJudgingSubmissions(ctx); err != nil {
		return nil, err
	}

	tx, err := lease.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin claim transaction: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		WITH picked AS (
			SELECT id
			FROM submission
			WHERE status = 'pending'
			ORDER BY submitted_at ASC, id ASC
			LIMIT ?
		)
		UPDATE submission
		SET status = 'judging'
		WHERE id IN (SELECT id FROM picked)
		RETURNING id;
	`, limit)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("claim submissions: %w", err)
	}

	ids := make([]int64, 0, limit)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return nil, fmt.Errorf("scan claimed submission id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		_ = tx.Rollback()
		return nil, fmt.Errorf("iterate claimed submission ids: %w", err)
	}
	if err := rows.Close(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("close claimed submission rows: %w", err)
	}

	if len(ids) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit empty claim transaction: %w", err)
		}
		return nil, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT
			s.id,
			s.source_code,
			q.reference_code,
			q.test_cases
		FROM submission s
		JOIN lesson_question lq
		  ON lq.id = s.lesson_question_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE s.id IN (%s)
		ORDER BY s.submitted_at ASC, s.id ASC
	`, strings.Join(placeholders, ", "))

	detailsRows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("query claimed submission details: %w", err)
	}
	defer detailsRows.Close()

	claimed := make([]claimedSubmission, 0, len(ids))
	for detailsRows.Next() {
		var item claimedSubmission
		if err := detailsRows.Scan(&item.ID, &item.SourceCode, &item.ReferenceCode, &item.TestCases); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("scan claimed submission details: %w", err)
		}
		claimed = append(claimed, item)
	}
	if err := detailsRows.Err(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("iterate claimed submission details: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim transaction: %w", err)
	}

	return claimed, nil
}

func (lease *lease) expireStaleJudgingSubmissions(ctx context.Context) error {
	_, err := lease.db.ExecContext(ctx, `
		UPDATE submission
		SET
			status = 'finished',
			verdict = 'system_error',
			error_message = 'judge timeout or interrupted',
			judge_report = '{"error":"judge timeout or interrupted"}',
			finished_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE status = 'judging'
		  AND submitted_at < strftime('%Y-%m-%dT%H:%M:%fZ', 'now', ?)
	`, fmt.Sprintf("-%d seconds", int(staleJudgingTimeout.Seconds())))
	if err != nil {
		return fmt.Errorf("expire stale judging submissions: %w", err)
	}
	return nil
}
