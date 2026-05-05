// Provides data access for inserting, querying, and updating submission results.

package submission

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"oj-lite/internal/platform/logger"
)

type repo struct {
	db  *sql.DB
	log *logger.Logger
}

func newRepo(database *sql.DB, log *logger.Logger) *repo {
	return &repo{
		db:  database,
		log: log,
	}
}

func (repo *repo) FindSubmissionTargetForStudent(
	ctx context.Context,
	classroomID, studentID, lessonQuestionID int64,
) (submissionTarget, error) {
	return scanSubmissionTarget(repo.db.QueryRowContext(ctx, `
		SELECT
			e.id,
			e.current_lesson_id,
			lq.id,
			q.id,
			q.title
		FROM enrollment e
		JOIN classroom c
		  ON c.id = e.classroom_id
		JOIN lesson_question lq
		  ON lq.id = ?
		 AND lq.lesson_id = c.current_lesson_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE e.classroom_id = ? AND e.student_id = ?
		LIMIT 1
	`, lessonQuestionID, classroomID, studentID))
}

func (repo *repo) CreateSubmission(
	ctx context.Context,
	target submissionTarget,
	sourceCode string,
) (Submission, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO submission (enrollment_id, lesson_id, lesson_question_id, status, source_code)
		VALUES (?, ?, ?, 'pending', ?)
	`, target.EnrollmentID, target.LessonID, target.LessonQuestionID, sourceCode)
	if err != nil {
		return Submission{}, err
	}

	submissionID, err := result.LastInsertId()
	if err != nil {
		return Submission{}, fmt.Errorf("read inserted submission id: %w", err)
	}

	return repo.FindSubmissionByIDAndEnrollment(ctx, target.EnrollmentID, submissionID)
}

func (repo *repo) FindEnrollmentIDForStudent(
	ctx context.Context,
	classroomID, studentID int64,
) (int64, error) {
	var enrollmentID int64
	err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM enrollment
		WHERE classroom_id = ? AND student_id = ?
		LIMIT 1
	`, classroomID, studentID).Scan(&enrollmentID)
	if err != nil {
		return 0, err
	}

	return enrollmentID, nil
}

func (repo *repo) ListSubmissionsByEnrollment(
	ctx context.Context,
	enrollmentID int64,
) ([]Submission, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.enrollment_id,
			s.lesson_id,
			s.lesson_question_id,
			q.id,
			q.title,
			s.status,
			COALESCE(s.verdict, ''),
			s.source_code,
			COALESCE(s.stdout_buffer, ''),
			COALESCE(s.error_message, ''),
			COALESCE(s.judge_report, ''),
			s.submitted_at,
			COALESCE(s.finished_at, '')
		FROM submission s
		JOIN lesson_question lq
		  ON lq.id = s.lesson_question_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE s.enrollment_id = ?
		ORDER BY s.submitted_at DESC, s.id DESC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Submission, 0)
	for rows.Next() {
		item, scanErr := scanSubmission(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (repo *repo) ListSubmissionsByEnrollmentAndLessonQuestion(
	ctx context.Context,
	enrollmentID, lessonQuestionID int64,
) ([]Submission, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.enrollment_id,
			s.lesson_id,
			s.lesson_question_id,
			q.id,
			q.title,
			s.status,
			COALESCE(s.verdict, ''),
			s.source_code,
			COALESCE(s.stdout_buffer, ''),
			COALESCE(s.error_message, ''),
			COALESCE(s.judge_report, ''),
			s.submitted_at,
			COALESCE(s.finished_at, '')
		FROM submission s
		JOIN lesson_question lq
		  ON lq.id = s.lesson_question_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE s.enrollment_id = ? AND s.lesson_question_id = ?
		ORDER BY s.submitted_at DESC, s.id DESC
	`, enrollmentID, lessonQuestionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Submission, 0)
	for rows.Next() {
		item, scanErr := scanSubmission(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (repo *repo) FindSubmissionByIDAndEnrollment(
	ctx context.Context,
	enrollmentID, submissionID int64,
) (Submission, error) {
	return scanSubmission(repo.db.QueryRowContext(ctx, `
		SELECT
			s.id,
			s.enrollment_id,
			s.lesson_id,
			s.lesson_question_id,
			q.id,
			q.title,
			s.status,
			COALESCE(s.verdict, ''),
			s.source_code,
			COALESCE(s.stdout_buffer, ''),
			COALESCE(s.error_message, ''),
			COALESCE(s.judge_report, ''),
			s.submitted_at,
			COALESCE(s.finished_at, '')
		FROM submission s
		JOIN lesson_question lq
		  ON lq.id = s.lesson_question_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE s.id = ? AND s.enrollment_id = ?
		LIMIT 1
	`, submissionID, enrollmentID))
}

func scanSubmissionTarget(scanner interface{ Scan(dest ...any) error }) (submissionTarget, error) {
	var target submissionTarget
	err := scanner.Scan(
		&target.EnrollmentID,
		&target.LessonID,
		&target.LessonQuestionID,
		&target.QuestionID,
		&target.QuestionTitle,
	)
	if err != nil {
		return submissionTarget{}, err
	}

	return target, nil
}

func scanSubmission(scanner interface{ Scan(dest ...any) error }) (Submission, error) {
	var (
		item         Submission
		stdoutBuffer string
		judgeReport  string
	)
	err := scanner.Scan(
		&item.ID,
		&item.EnrollmentID,
		&item.LessonID,
		&item.LessonQuestionID,
		&item.QuestionID,
		&item.QuestionTitle,
		&item.Status,
		&item.Verdict,
		&item.SourceCode,
		&stdoutBuffer,
		&item.ErrorMessage,
		&judgeReport,
		&item.SubmittedAt,
		&item.FinishedAt,
	)
	if err != nil {
		return Submission{}, err
	}

	item.StdoutBuffer = rawJSON(stdoutBuffer)
	item.JudgeReport = rawJSON(judgeReport)
	return item, nil
}

func rawJSON(value string) json.RawMessage {
	if value == "" {
		return nil
	}
	return json.RawMessage(value)
}
