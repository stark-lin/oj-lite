package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"oj-lite/internal/platform/config"
	pdb "oj-lite/internal/platform/db"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/seed"
)

func TestLeaseClaimPendingSubmissionsMarksJudgingAndReturnsDetails(t *testing.T) {
	database := openSchedulerTestDB(t)
	seedSubmissionFixture(t, database, `
function solution(a, b)
  return a + b
end
`, `
function solution(a, b)
  return a + b
end
`, `[{"input":[1,2]}]`)

	lease := newLease(database, logger.NewLogger("scheduler-test"))
	claimed, err := lease.claimPendingSubmissions(context.Background(), 4)
	if err != nil {
		t.Fatalf("claim pending submissions: %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected 1 claimed submission, got %d", len(claimed))
	}

	var status string
	if err := database.QueryRowContext(context.Background(), `SELECT status FROM submission WHERE id = ?`, claimed[0].ID).Scan(&status); err != nil {
		t.Fatalf("query updated submission: %v", err)
	}
	if status != "judging" {
		t.Fatalf("expected claimed submission status=judging, got %q", status)
	}
}

func TestLeaseClaimPendingSubmissionsExpiresStaleJudgingSubmissions(t *testing.T) {
	database := openSchedulerTestDB(t)
	staleID := seedSubmissionFixture(t, database, `
function solution(a)
  return a
end
`, `
function solution(a)
  return a
end
`, `[{"input":[1]}]`)
	freshID := seedSubmissionFixture(t, database, `
function solution(a)
  return a
end
`, `
function solution(a)
  return a
end
`, `[{"input":[1]}]`)

	if _, err := database.ExecContext(context.Background(), `
		UPDATE submission
		SET status = 'judging',
		    submitted_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now', '-2 minutes')
		WHERE id = ?
	`, staleID); err != nil {
		t.Fatalf("mark stale submission judging: %v", err)
	}
	if _, err := database.ExecContext(context.Background(), `
		UPDATE submission
		SET status = 'judging'
		WHERE id = ?
	`, freshID); err != nil {
		t.Fatalf("mark fresh submission judging: %v", err)
	}

	lease := newLease(database, logger.NewLogger("scheduler-test"))
	if _, err := lease.claimPendingSubmissions(context.Background(), 1); err != nil {
		t.Fatalf("claim pending submissions: %v", err)
	}

	var staleStatus string
	var staleVerdict string
	var staleErrorMessage string
	var staleJudgeReport string
	if err := database.QueryRowContext(context.Background(), `
		SELECT status, verdict, error_message, judge_report
		FROM submission
		WHERE id = ?
	`, staleID).Scan(&staleStatus, &staleVerdict, &staleErrorMessage, &staleJudgeReport); err != nil {
		t.Fatalf("query stale submission: %v", err)
	}
	if staleStatus != "finished" || staleVerdict != "system_error" {
		t.Fatalf("stale submission status/verdict = %q/%q, want finished/system_error", staleStatus, staleVerdict)
	}
	if staleErrorMessage != "judge timeout or interrupted" {
		t.Fatalf("stale submission error_message = %q", staleErrorMessage)
	}
	if staleJudgeReport != `{"error":"judge timeout or interrupted"}` {
		t.Fatalf("stale submission judge_report = %q", staleJudgeReport)
	}

	var freshStatus string
	var freshVerdict sql.NullString
	if err := database.QueryRowContext(context.Background(), `
		SELECT status, verdict
		FROM submission
		WHERE id = ?
	`, freshID).Scan(&freshStatus, &freshVerdict); err != nil {
		t.Fatalf("query fresh submission: %v", err)
	}
	if freshStatus != "judging" || freshVerdict.Valid {
		t.Fatalf("fresh submission status/verdict = %q/%#v, want judging/NULL", freshStatus, freshVerdict)
	}
}

func TestWorkerProcessClaimedScriptWritesFinishedResult(t *testing.T) {
	database := openSchedulerTestDB(t)
	submissionID := seedSubmissionFixture(t, database, `
function solution(a, b)
  return a + b
end
`, `
function solution(a, b)
  print(a, b)
  return a + b
end
`, `[{"input":[1,2]},{"input":[3,4]}]`)

	lease := newLease(database, logger.NewLogger("scheduler-test"))
	claimed, err := lease.claimPendingSubmissions(context.Background(), 1)
	if err != nil {
		t.Fatalf("claim pending submissions: %v", err)
	}

	worker := newWorker(database, logger.NewLogger("scheduler-test"))
	if err := worker.processClaimedScript(context.Background(), claimed[0]); err != nil {
		t.Fatalf("process claimed script: %v", err)
	}

	var (
		status       string
		verdict      string
		stdoutBuffer sql.NullString
		errorMessage sql.NullString
		judgeReport  sql.NullString
		finishedAt   sql.NullString
	)
	if err := database.QueryRowContext(context.Background(), `
		SELECT status, verdict, stdout_buffer, error_message, judge_report, finished_at
		FROM submission
		WHERE id = ?
	`, submissionID).Scan(&status, &verdict, &stdoutBuffer, &errorMessage, &judgeReport, &finishedAt); err != nil {
		t.Fatalf("query finished submission: %v", err)
	}

	if status != "finished" {
		t.Fatalf("expected finished status, got %q", status)
	}
	if verdict != "accepted" {
		t.Fatalf("expected accepted verdict, got %q", verdict)
	}
	if !stdoutBuffer.Valid {
		t.Fatalf("unexpected stdout_buffer: %#v", stdoutBuffer)
	}

	var stdout stdoutBufferShape
	if err := json.Unmarshal([]byte(stdoutBuffer.String), &stdout); err != nil {
		t.Fatalf("unmarshal stdout_buffer: %v", err)
	}
	if len(stdout.Cases) != 2 {
		t.Fatalf("expected 2 stdout cases, got %#v", stdout)
	}
	if stdout.Cases[0].Index != 1 || stdout.Cases[0].Stdout != "1\t2" {
		t.Fatalf("unexpected first stdout case: %#v", stdout.Cases[0])
	}
	if stdout.Cases[1].Index != 2 || stdout.Cases[1].Stdout != "3\t4" {
		t.Fatalf("unexpected second stdout case: %#v", stdout.Cases[1])
	}
	if errorMessage.Valid {
		t.Fatalf("expected no error_message, got %#v", errorMessage)
	}
	if !judgeReport.Valid || judgeReport.String == "" {
		t.Fatal("expected judge_report JSON to be written")
	}
	if !finishedAt.Valid || finishedAt.String == "" {
		t.Fatal("expected finished_at to be set")
	}

	var report judgeReportShape
	if err := json.Unmarshal([]byte(judgeReport.String), &report); err != nil {
		t.Fatalf("unmarshal judge report: %v", err)
	}
	if len(report.Cases) != 2 {
		t.Fatalf("expected 2 judge report cases, got %d", len(report.Cases))
	}
}

func TestWorkerProcessClaimedScriptTreatsAggregateStdoutLimitAsRuntimeError(t *testing.T) {
	database := openSchedulerTestDB(t)
	longLine := strings.Repeat("x", 700)
	testCases := make([]map[string][]int, 12)
	for index := range testCases {
		testCases[index] = map[string][]int{"input": []int{index + 1}}
	}
	rawTestCases, err := json.Marshal(testCases)
	if err != nil {
		t.Fatalf("marshal test cases: %v", err)
	}

	seedSubmissionFixture(t, database, `
function solution(a)
  return a
end
`, fmt.Sprintf(`
function solution(a)
  print(%q)
  return a
end
`, longLine), string(rawTestCases))

	lease := newLease(database, logger.NewLogger("scheduler-test"))
	claimed, err := lease.claimPendingSubmissions(context.Background(), 1)
	if err != nil {
		t.Fatalf("claim pending submissions: %v", err)
	}

	worker := newWorker(database, logger.NewLogger("scheduler-test"))
	if err := worker.processClaimedScript(context.Background(), claimed[0]); err != nil {
		t.Fatalf("process claimed script: %v", err)
	}

	var verdict string
	var stdoutBuffer sql.NullString
	var errorMessage sql.NullString
	var judgeReport sql.NullString
	if err := database.QueryRowContext(context.Background(), `
		SELECT verdict, stdout_buffer, error_message, judge_report
		FROM submission
		WHERE id = ?
	`, claimed[0].ID).Scan(&verdict, &stdoutBuffer, &errorMessage, &judgeReport); err != nil {
		t.Fatalf("query stdout limit submission: %v", err)
	}

	if verdict != "runtime_error" {
		t.Fatalf("expected runtime_error verdict, got %q", verdict)
	}
	if !errorMessage.Valid || errorMessage.String != "STDOUT LIMIT EXCEEDED" {
		t.Fatalf("expected stdout limit error_message, got %#v", errorMessage)
	}
	if !stdoutBuffer.Valid {
		t.Fatal("expected stdout_buffer to be written")
	}
	if len([]byte(stdoutBuffer.String)) > stdoutBufferMaxBytes {
		t.Fatalf("stdout_buffer exceeds limit: %d", len([]byte(stdoutBuffer.String)))
	}

	var stdout stdoutBufferShape
	if err := json.Unmarshal([]byte(stdoutBuffer.String), &stdout); err != nil {
		t.Fatalf("unmarshal stdout_buffer: %v", err)
	}
	if len(stdout.Cases) != 1 || stdout.Cases[0].Stdout != "STDOUT LIMIT EXCEEDED" {
		t.Fatalf("unexpected stdout limit buffer: %#v", stdout)
	}

	var report struct {
		Cases []struct {
			Student struct {
				ErrorMessage string `json:"errorMessage"`
				StdoutBuffer string `json:"stdoutBuffer"`
			} `json:"student"`
		} `json:"cases"`
	}
	if !judgeReport.Valid || judgeReport.String == "" {
		t.Fatal("expected judge_report JSON to be written")
	}
	if err := json.Unmarshal([]byte(judgeReport.String), &report); err != nil {
		t.Fatalf("unmarshal judge report: %v", err)
	}
	if len(report.Cases) == 0 || report.Cases[0].Student.ErrorMessage != "STDOUT LIMIT EXCEEDED" || report.Cases[0].Student.StdoutBuffer != "STDOUT LIMIT EXCEEDED" {
		t.Fatalf("unexpected judge report stdout limit case: %#v", report.Cases)
	}
}

func TestWorkerProcessClaimedScriptWritesRuntimeError(t *testing.T) {
	database := openSchedulerTestDB(t)
	seedSubmissionFixture(t, database, `
function solution(a)
  return a + 1
end
`, `
function solution(a)
  print("before")
  local x = nil
  return x.value
end
`, `[{"input":[1]}]`)

	lease := newLease(database, logger.NewLogger("scheduler-test"))
	claimed, err := lease.claimPendingSubmissions(context.Background(), 1)
	if err != nil {
		t.Fatalf("claim pending submissions: %v", err)
	}

	worker := newWorker(database, logger.NewLogger("scheduler-test"))
	if err := worker.processClaimedScript(context.Background(), claimed[0]); err != nil {
		t.Fatalf("process claimed script: %v", err)
	}

	var verdict string
	var stdoutBuffer sql.NullString
	var errorMessage sql.NullString
	if err := database.QueryRowContext(context.Background(), `
		SELECT verdict, stdout_buffer, error_message
		FROM submission
		WHERE id = ?
	`, claimed[0].ID).Scan(&verdict, &stdoutBuffer, &errorMessage); err != nil {
		t.Fatalf("query runtime_error submission: %v", err)
	}

	if verdict != "runtime_error" {
		t.Fatalf("expected runtime_error verdict, got %q", verdict)
	}
	if !stdoutBuffer.Valid {
		t.Fatalf("unexpected stdout_buffer: %#v", stdoutBuffer)
	}

	var stdout stdoutBufferShape
	if err := json.Unmarshal([]byte(stdoutBuffer.String), &stdout); err != nil {
		t.Fatalf("unmarshal stdout_buffer: %v", err)
	}
	if len(stdout.Cases) != 1 {
		t.Fatalf("expected 1 stdout case, got %#v", stdout)
	}
	if !strings.Contains(stdout.Cases[0].Stdout, "before") || !strings.Contains(stdout.Cases[0].Stdout, "[runtime_error]") {
		t.Fatalf("unexpected runtime_error stdout case: %#v", stdout.Cases[0])
	}
	if !errorMessage.Valid || errorMessage.String == "" {
		t.Fatal("expected error_message to be written")
	}
}

type stdoutBufferShape struct {
	Cases []struct {
		Index  int    `json:"index"`
		Stdout string `json:"stdout"`
	} `json:"cases"`
}

type judgeReportShape struct {
	Cases []struct{} `json:"cases"`
}

func openSchedulerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := pdb.Open(context.Background(), config.DBConfig{
		Path:        filepath.Join(t.TempDir(), "oj-lite.db"),
		BusyTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	if err := pdb.Migrate(context.Background(), database); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	if err := seed.SeedDemoAccounts(context.Background(), database); err != nil {
		t.Fatalf("seed demo data: %v", err)
	}

	return database
}

func seedSubmissionFixture(t *testing.T, database *sql.DB, referenceCode, sourceCode, testCases string) int64 {
	t.Helper()

	tx, err := database.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin fixture transaction: %v", err)
	}
	defer tx.Rollback()

	var enrollmentID int64
	var classroomID int64
	if err := tx.QueryRowContext(context.Background(), `SELECT id FROM enrollment LIMIT 1`).Scan(&enrollmentID); err != nil {
		t.Fatalf("load enrollment id: %v", err)
	}
	if err := tx.QueryRowContext(context.Background(), `SELECT classroom_id FROM enrollment WHERE id = ?`, enrollmentID).Scan(&classroomID); err != nil {
		t.Fatalf("load classroom id: %v", err)
	}

	lessonResult, err := tx.ExecContext(context.Background(), `
		INSERT INTO lesson (title, description, sort_order)
		VALUES ('lesson', 'lesson', ?)
	`, time.Now().UnixNano())
	if err != nil {
		t.Fatalf("insert lesson: %v", err)
	}
	lessonID, err := lessonResult.LastInsertId()
	if err != nil {
		t.Fatalf("lesson last insert id: %v", err)
	}

	questionResult, err := tx.ExecContext(context.Background(), `
		INSERT INTO question (title, description, starter_code, reference_code, test_cases)
		VALUES ('question', 'question', '', ?, ?)
	`, referenceCode, testCases)
	if err != nil {
		t.Fatalf("insert question: %v", err)
	}
	questionID, err := questionResult.LastInsertId()
	if err != nil {
		t.Fatalf("question last insert id: %v", err)
	}

	lessonQuestionResult, err := tx.ExecContext(context.Background(), `
		INSERT INTO lesson_question (lesson_id, question_id, sort_order)
		VALUES (?, ?, 1)
	`, lessonID, questionID)
	if err != nil {
		t.Fatalf("insert lesson_question: %v", err)
	}
	lessonQuestionID, err := lessonQuestionResult.LastInsertId()
	if err != nil {
		t.Fatalf("lesson_question last insert id: %v", err)
	}

	if _, err := tx.ExecContext(context.Background(), `
		UPDATE classroom
		SET current_lesson_id = ?
		WHERE id = ?
	`, lessonID, classroomID); err != nil {
		t.Fatalf("update classroom current lesson: %v", err)
	}

	if _, err := tx.ExecContext(context.Background(), `
		UPDATE enrollment
		SET current_lesson_id = ?
		WHERE id = ?
	`, lessonID, enrollmentID); err != nil {
		t.Fatalf("update enrollment current lesson: %v", err)
	}

	submissionResult, err := tx.ExecContext(context.Background(), `
		INSERT INTO submission (enrollment_id, lesson_id, lesson_question_id, status, source_code)
		VALUES (?, ?, ?, 'pending', ?)
	`, enrollmentID, lessonID, lessonQuestionID, sourceCode)
	if err != nil {
		t.Fatalf("insert submission: %v", err)
	}
	submissionID, err := submissionResult.LastInsertId()
	if err != nil {
		t.Fatalf("submission last insert id: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit fixture transaction: %v", err)
	}

	return submissionID
}
