// Provides data access for questions.

package question

import (
	"context"
	"database/sql"
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

func (repo *repo) CreateQuestion(
	ctx context.Context,
	title, description, starterCode, referenceCode, testCases string,
) (Question, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO question (title, description, starter_code, reference_code, test_cases)
		VALUES (?, ?, ?, ?, ?)
	`, title, description, starterCode, referenceCode, testCases)
	if err != nil {
		return Question{}, err
	}

	questionID, err := result.LastInsertId()
	if err != nil {
		return Question{}, fmt.Errorf("read inserted question id: %w", err)
	}

	return repo.FindQuestionByID(ctx, questionID)
}

func (repo *repo) ListQuestions(ctx context.Context) ([]Question, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, title, description, starter_code, reference_code, test_cases, created_at
		FROM question
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []Question
	for rows.Next() {
		question, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func (repo *repo) FindQuestionByID(ctx context.Context, questionID int64) (Question, error) {
	return scanQuestion(repo.db.QueryRowContext(ctx, `
		SELECT id, title, description, starter_code, reference_code, test_cases, created_at
		FROM question
		WHERE id = ?
		LIMIT 1
	`, questionID))
}

func (repo *repo) FindStudentQuestionByLessonQuestionID(
	ctx context.Context,
	classroomID, studentID, lessonQuestionID int64,
) (StudentQuestion, error) {
	return scanStudentQuestion(repo.db.QueryRowContext(ctx, `
		SELECT q.id, lq.id, q.title, q.description, q.starter_code, lq.sort_order, q.created_at
		FROM enrollment e
		JOIN classroom c
		  ON c.id = e.classroom_id
		JOIN lesson_question lq
		  ON lq.lesson_id = c.current_lesson_id
		JOIN question q
		  ON q.id = lq.question_id
		WHERE e.classroom_id = ? AND e.student_id = ? AND lq.id = ?
		LIMIT 1
	`, classroomID, studentID, lessonQuestionID))
}

func scanQuestion(scanner interface{ Scan(dest ...any) error }) (Question, error) {
	var question Question
	err := scanner.Scan(
		&question.ID,
		&question.Title,
		&question.Description,
		&question.StarterCode,
		&question.ReferenceCode,
		&question.TestCases,
		&question.CreatedAt,
	)
	if err != nil {
		return Question{}, err
	}

	return question, nil
}

func scanStudentQuestion(scanner interface{ Scan(dest ...any) error }) (StudentQuestion, error) {
	var question StudentQuestion
	err := scanner.Scan(
		&question.ID,
		&question.LessonQuestionID,
		&question.Title,
		&question.Description,
		&question.StarterCode,
		&question.SortOrder,
		&question.CreatedAt,
	)
	if err != nil {
		return StudentQuestion{}, err
	}

	return question, nil
}
