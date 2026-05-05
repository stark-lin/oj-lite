// Processes one submission end to end, including claiming, judging, and writing back results.

package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"oj-lite/internal/judge"
	"oj-lite/internal/platform/logger"
)

const stdoutBufferMaxBytes = 8192

type worker struct {
	db    *sql.DB
	judge *judge.Engine
	log   *logger.Logger
}

func newWorker(database *sql.DB, log *logger.Logger) *worker {
	return &worker{
		db:    database,
		judge: judge.New(log),
		log:   log,
	}
}

func (worker *worker) processClaimedScript(ctx context.Context, script claimedSubmission) error {
	report, err := worker.judge.Run(ctx, judge.Request{
		ReferenceCode: script.ReferenceCode,
		SourceCode:    script.SourceCode,
		TestCases:     script.TestCases,
	})
	if err != nil {
		return worker.finishWithSystemError(ctx, script.ID, err)
	}
	limitSubmissionStdout(&report)

	payload, err := json.Marshal(report)
	if err != nil {
		return worker.finishWithSystemError(ctx, script.ID, fmt.Errorf("marshal judge report: %w", err))
	}

	stdoutBuffer, err := aggregateStdout(report)
	if err != nil {
		return worker.finishWithSystemError(ctx, script.ID, fmt.Errorf("marshal stdout buffer: %w", err))
	}
	errorMessage, verdict := deriveOutcome(report)
	if err := worker.finishSubmission(ctx, script.ID, verdict, stdoutBuffer, errorMessage, string(payload)); err != nil {
		return fmt.Errorf("finish submission %d: %w", script.ID, err)
	}

	return nil
}

func limitSubmissionStdout(report *judge.Report) {
	stdoutBuffer, err := marshalStdoutReport(*report)
	if err == nil && len([]byte(stdoutBuffer)) <= stdoutBufferMaxBytes {
		return
	}

	for index := range report.Cases {
		if report.Cases[index].Student.StdoutBuffer == "" && !report.Cases[index].Student.StdoutLimitExceeded {
			continue
		}
		report.Cases[index].Student.StdoutBuffer = judge.StdoutLimitExceededMessage
		report.Cases[index].Student.ErrorMessage = judge.StdoutLimitExceededMessage
		report.Cases[index].Student.ReturnValues = nil
		report.Cases[index].Student.StdoutLimitExceeded = true
		report.Cases[index].Comparison = judge.ComparisonReport{
			Matched: false,
			Reason:  fmt.Sprintf("student execution failed: %s", judge.StdoutLimitExceededMessage),
		}
	}
}

func (worker *worker) finishWithSystemError(ctx context.Context, submissionID int64, cause error) error {
	if err := worker.finishSubmission(ctx, submissionID, "system_error", "", cause.Error(), ""); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (worker *worker) finishSubmission(ctx context.Context, submissionID int64, verdict, stdoutBuffer, errorMessage, judgeReport string) error {
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	var stdoutArg any
	if stdoutBuffer != "" {
		stdoutArg = stdoutBuffer
	}

	var errorArg any
	if errorMessage != "" {
		errorArg = errorMessage
	}

	var reportArg any
	if judgeReport != "" {
		reportArg = judgeReport
	}

	result, err := worker.db.ExecContext(writeCtx, `
		UPDATE submission
		SET
			status = 'finished',
			verdict = ?,
			stdout_buffer = ?,
			error_message = ?,
			judge_report = ?,
			finished_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?
		  AND status = 'judging'
	`, verdict, stdoutArg, errorArg, reportArg, submissionID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("expected to finish exactly one submission, got %d", rowsAffected)
	}

	return nil
}

func aggregateStdout(report judge.Report) (string, error) {
	if hasStdoutLimitExceeded(report) {
		return marshalStdoutLimitExceeded(report)
	}

	raw, err := marshalStdoutReport(report)
	if err != nil || raw == "" {
		return raw, err
	}
	if len([]byte(raw)) <= stdoutBufferMaxBytes {
		return raw, nil
	}

	return marshalStdoutLimitExceeded(report)
}

func hasStdoutLimitExceeded(report judge.Report) bool {
	for _, item := range report.Cases {
		if item.Student.StdoutLimitExceeded {
			return true
		}
	}
	return false
}

func marshalStdoutReport(report judge.Report) (string, error) {
	stdout := judge.StdoutReport{
		Cases: make([]judge.StdoutCase, 0, len(report.Cases)),
	}
	for _, item := range report.Cases {
		if item.Student.StdoutBuffer == "" {
			continue
		}
		stdout.Cases = append(stdout.Cases, judge.StdoutCase{
			Index:  item.Index,
			Stdout: item.Student.StdoutBuffer,
		})
	}
	if len(stdout.Cases) == 0 {
		return "", nil
	}

	raw, err := json.Marshal(stdout)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func marshalStdoutLimitExceeded(report judge.Report) (string, error) {
	caseIndex := 1
	for _, item := range report.Cases {
		if item.Student.StdoutLimitExceeded {
			caseIndex = item.Index
			break
		}
	}
	if caseIndex == 1 {
		for _, item := range report.Cases {
			if item.Student.StdoutBuffer != "" {
				caseIndex = item.Index
				break
			}
		}
	}

	raw, err := json.Marshal(judge.StdoutReport{
		Cases: []judge.StdoutCase{{
			Index:  caseIndex,
			Stdout: judge.StdoutLimitExceededMessage,
		}},
	})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func deriveOutcome(report judge.Report) (string, string) {
	for _, item := range report.Cases {
		if item.Reference.ErrorMessage != "" {
			return item.Reference.ErrorMessage, "system_error"
		}
	}
	for _, item := range report.Cases {
		if item.Student.ErrorMessage != "" {
			return item.Student.ErrorMessage, "runtime_error"
		}
	}
	for _, item := range report.Cases {
		if !item.Comparison.Matched {
			return "", "wrong_answer"
		}
	}
	return "", "accepted"
}
