// Provides data access for lessons and lesson-question relationships.

package lesson

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

func (repo *repo) CreateLesson(ctx context.Context, title, description string, sortOrder int) (Lesson, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO lesson (title, description, sort_order)
		VALUES (?, ?, ?)
	`, title, description, sortOrder)
	if err != nil {
		return Lesson{}, err
	}

	lessonID, err := result.LastInsertId()
	if err != nil {
		return Lesson{}, fmt.Errorf("read inserted lesson id: %w", err)
	}

	return repo.FindLessonByID(ctx, lessonID)
}

func (repo *repo) ListLessons(ctx context.Context) ([]Lesson, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, title, description, sort_order, created_at
		FROM lesson
		ORDER BY sort_order ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		lesson, err := scanLesson(rows)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (repo *repo) FindLessonByID(ctx context.Context, lessonID int64) (Lesson, error) {
	return scanLesson(repo.db.QueryRowContext(ctx, `
		SELECT id, title, description, sort_order, created_at
		FROM lesson
		WHERE id = ?
		LIMIT 1
	`, lessonID))
}

func (repo *repo) ExistsLessonSortOrder(ctx context.Context, sortOrder int) (bool, error) {
	var exists bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM lesson
			WHERE sort_order = ?
		)
	`, sortOrder).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repo *repo) ListLessonQuestions(ctx context.Context, lessonID int64) ([]LessonQuestion, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT lq.id, lq.lesson_id, q.id, q.title, lq.sort_order, lq.created_at
		FROM lesson_question lq
		JOIN question q
		  ON q.id = lq.question_id
		WHERE lq.lesson_id = ?
		ORDER BY lq.sort_order ASC, lq.id ASC
	`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]LessonQuestion, 0)
	for rows.Next() {
		item, scanErr := scanLessonQuestion(rows)
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

func scanLesson(scanner interface{ Scan(dest ...any) error }) (Lesson, error) {
	var lesson Lesson
	err := scanner.Scan(
		&lesson.ID,
		&lesson.Title,
		&lesson.Description,
		&lesson.SortOrder,
		&lesson.CreatedAt,
	)
	if err != nil {
		return Lesson{}, err
	}

	return lesson, nil
}

func scanLessonQuestion(scanner interface{ Scan(dest ...any) error }) (LessonQuestion, error) {
	var question LessonQuestion
	err := scanner.Scan(
		&question.ID,
		&question.LessonID,
		&question.QuestionID,
		&question.Title,
		&question.SortOrder,
		&question.CreatedAt,
	)
	if err != nil {
		return LessonQuestion{}, err
	}

	return question, nil
}
