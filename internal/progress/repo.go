// Queries submission and completion aggregates by classroom.

package progress

import (
	"context"
	"database/sql"
	"encoding/json"

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

func (repo *repo) ClassroomExistsForTeacher(ctx context.Context, classroomID, teacherID int64) (bool, error) {
	var exists bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM classroom
			WHERE id = ? AND teacher_id = ?
		)
	`, classroomID, teacherID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repo *repo) ListStudentProgressByClassroom(ctx context.Context, classroomID, teacherID int64) ([]StudentProgress, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			u.id,
			u.username,
			u.role,
			u.status,
			u.created_at,
			COALESCE(l.id, 0),
			COALESCE(l.title, ''),
			COALESCE((
				SELECT COUNT(*)
				FROM lesson_question lq
				WHERE lq.lesson_id = c.current_lesson_id
			), 0),
			COALESCE((
				SELECT COUNT(DISTINCT s.lesson_question_id)
				FROM submission s
				WHERE s.enrollment_id = e.id
				  AND s.lesson_id = c.current_lesson_id
				  AND s.verdict = 'accepted'
			), 0),
			latest.id,
			latest.enrollment_id,
			COALESCE(latest.lesson_id, 0),
			COALESCE(latest.lesson_question_id, 0),
			COALESCE(q.id, 0),
			COALESCE(q.title, ''),
			COALESCE(latest.status, ''),
			COALESCE(latest.verdict, ''),
			COALESCE(latest.stdout_buffer, ''),
			COALESCE(latest.error_message, ''),
			COALESCE(latest.judge_report, ''),
			COALESCE(latest.submitted_at, ''),
			COALESCE(latest.finished_at, '')
		FROM classroom c
		JOIN enrollment e
		  ON e.classroom_id = c.id
		JOIN user_account u
		  ON u.id = e.student_id
		LEFT JOIN lesson l
		  ON l.id = c.current_lesson_id
		LEFT JOIN submission latest
		  ON latest.id = (
			  SELECT s2.id
			  FROM submission s2
			  WHERE s2.enrollment_id = e.id
			  ORDER BY s2.submitted_at DESC, s2.id DESC
			  LIMIT 1
		  )
		LEFT JOIN lesson_question lq_latest
		  ON lq_latest.id = latest.lesson_question_id
		LEFT JOIN question q
		  ON q.id = lq_latest.question_id
		WHERE c.id = ? AND c.teacher_id = ?
		ORDER BY u.id ASC
	`, classroomID, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StudentProgress, 0)
	for rows.Next() {
		item, scanErr := scanStudentProgress(rows)
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

func (repo *repo) ListSubmissionsByClassroom(ctx context.Context, classroomID, teacherID int64) ([]Submission, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			s.enrollment_id,
			u.id,
			u.username,
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
		FROM classroom c
		JOIN enrollment e
		  ON e.classroom_id = c.id
		JOIN submission s
		  ON s.enrollment_id = e.id
		JOIN user_account u
		  ON u.id = e.student_id
		JOIN lesson_question lq
		  ON lq.id = s.lesson_question_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE c.id = ? AND c.teacher_id = ?
		  AND s.lesson_id = c.current_lesson_id
		ORDER BY s.submitted_at DESC, s.id DESC
	`, classroomID, teacherID)
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

func (repo *repo) FindSubmissionByClassroom(ctx context.Context, classroomID, teacherID, submissionID int64) (Submission, error) {
	return scanSubmission(repo.db.QueryRowContext(ctx, `
		SELECT
			s.id,
			s.enrollment_id,
			u.id,
			u.username,
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
		FROM classroom c
		JOIN enrollment e
		  ON e.classroom_id = c.id
		JOIN submission s
		  ON s.enrollment_id = e.id
		JOIN user_account u
		  ON u.id = e.student_id
		JOIN lesson_question lq
		  ON lq.id = s.lesson_question_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE c.id = ? AND c.teacher_id = ? AND s.id = ?
		LIMIT 1
	`, classroomID, teacherID, submissionID))
}

func (repo *repo) DeleteSubmissionByClassroom(ctx context.Context, classroomID, teacherID, submissionID int64) error {
	result, err := repo.db.ExecContext(ctx, `
		DELETE FROM submission
		WHERE id = ?
		  AND EXISTS (
			  SELECT 1
			  FROM enrollment e
			  JOIN classroom c
			    ON c.id = e.classroom_id
			  WHERE e.id = submission.enrollment_id
			    AND c.id = ?
			    AND c.teacher_id = ?
		  )
	`, submissionID, classroomID, teacherID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func scanStudentProgress(scanner interface{ Scan(dest ...any) error }) (StudentProgress, error) {
	var (
		item                StudentProgress
		latestID            sql.NullInt64
		latestEnrollmentID  sql.NullInt64
		latestLessonID      int64
		latestLessonQID     int64
		latestQuestionID    int64
		latestQuestionTitle string
		latestStatus        string
		latestVerdict       string
		latestStdout        string
		latestError         string
		latestReport        string
		latestSubmittedAt   string
		latestFinishedAt    string
	)

	err := scanner.Scan(
		&item.StudentID,
		&item.Username,
		&item.Role,
		&item.Status,
		&item.CreatedAt,
		&item.CurrentLessonID,
		&item.CurrentLessonTitle,
		&item.TotalCount,
		&item.AcceptedCount,
		&latestID,
		&latestEnrollmentID,
		&latestLessonID,
		&latestLessonQID,
		&latestQuestionID,
		&latestQuestionTitle,
		&latestStatus,
		&latestVerdict,
		&latestStdout,
		&latestError,
		&latestReport,
		&latestSubmittedAt,
		&latestFinishedAt,
	)
	if err != nil {
		return StudentProgress{}, err
	}

	if latestID.Valid {
		item.LatestSubmission = &Submission{
			ID:               latestID.Int64,
			EnrollmentID:     latestEnrollmentID.Int64,
			StudentID:        item.StudentID,
			StudentUsername:  item.Username,
			LessonID:         latestLessonID,
			LessonQuestionID: latestLessonQID,
			QuestionID:       latestQuestionID,
			QuestionTitle:    latestQuestionTitle,
			Status:           latestStatus,
			Verdict:          latestVerdict,
			StdoutBuffer:     rawJSON(latestStdout),
			ErrorMessage:     latestError,
			JudgeReport:      rawJSON(latestReport),
			SubmittedAt:      latestSubmittedAt,
			FinishedAt:       latestFinishedAt,
		}
	}

	return item, nil
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
		&item.StudentID,
		&item.StudentUsername,
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
