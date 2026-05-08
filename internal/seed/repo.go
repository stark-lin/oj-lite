package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"oj-lite/internal/platform/password"
	"oj-lite/internal/platform/user"
)

type repo struct {
	db *sql.DB
}

func newRepo(database *sql.DB) *repo {
	return &repo{
		db: database,
	}
}

type seedTx struct {
	tx *sql.Tx
}

func (repo *repo) withTx(ctx context.Context, fn func(*seedTx) error) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin demo seed transaction: %w", err)
	}

	if err := fn(&seedTx{tx: tx}); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit demo seed transaction: %w", err)
	}

	return nil
}

func (seed *seedTx) ensureDemoUser(ctx context.Context, username, rawPassword, role string) (int64, error) {
	var accountID int64
	var existingRole string
	err := seed.tx.QueryRowContext(ctx, `
		SELECT id, role
		FROM user_account
		WHERE username = ?
		LIMIT 1
	`, username).Scan(&accountID, &existingRole)
	if err == nil {
		if existingRole != role {
			return 0, fmt.Errorf("demo username %q already exists with role %q", username, existingRole)
		}

		return accountID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("find demo user %q: %w", username, err)
	}

	passwordHash, err := password.Hash(rawPassword)
	if err != nil {
		return 0, fmt.Errorf("hash demo password for %q: %w", username, err)
	}

	result, err := seed.tx.ExecContext(ctx, `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES (?, ?, ?, ?)
	`, username, passwordHash, role, user.StatusActive)
	if err != nil {
		return 0, fmt.Errorf("insert demo user %q: %w", username, err)
	}

	accountID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted demo user id for %q: %w", username, err)
	}

	return accountID, nil
}

func (seed *seedTx) ensureDemoClassroom(ctx context.Context, teacherID int64, name string) (int64, error) {
	var classroomID int64
	err := seed.tx.QueryRowContext(ctx, `
		SELECT id
		FROM classroom
		WHERE teacher_id = ? AND name = ?
		LIMIT 1
	`, teacherID, name).Scan(&classroomID)
	if err == nil {
		return classroomID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("find demo classroom: %w", err)
	}

	result, err := seed.tx.ExecContext(ctx, `
		INSERT INTO classroom (teacher_id, name)
		VALUES (?, ?)
	`, teacherID, name)
	if err != nil {
		return 0, fmt.Errorf("insert demo classroom: %w", err)
	}

	classroomID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted demo classroom id: %w", err)
	}

	return classroomID, nil
}

func (seed *seedTx) ensureEmbeddedLesson(ctx context.Context, lesson lessonSeed) (int64, error) {
	var lessonID int64
	err := seed.tx.QueryRowContext(ctx, `
		SELECT id
		FROM lesson
		WHERE sort_order = ?
		LIMIT 1
	`, lesson.SortOrder).Scan(&lessonID)
	if err == nil {
		if _, err := seed.tx.ExecContext(ctx, `
			UPDATE lesson
			SET title = ?, description = ?
			WHERE id = ?
		`, lesson.Title, lesson.Description, lessonID); err != nil {
			return 0, fmt.Errorf("update embedded lesson %d: %w", lesson.SortOrder, err)
		}

		for _, question := range lesson.Questions {
			if err := seed.ensureEmbeddedQuestion(ctx, lessonID, question); err != nil {
				return 0, err
			}
		}

		return lessonID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("find embedded lesson %d: %w", lesson.SortOrder, err)
	}

	result, err := seed.tx.ExecContext(ctx, `
		INSERT INTO lesson (title, description, sort_order)
		VALUES (?, ?, ?)
	`, lesson.Title, lesson.Description, lesson.SortOrder)
	if err != nil {
		return 0, fmt.Errorf("insert embedded lesson %d: %w", lesson.SortOrder, err)
	}

	lessonID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted embedded lesson %d id: %w", lesson.SortOrder, err)
	}

	for _, question := range lesson.Questions {
		if err := seed.ensureEmbeddedQuestion(ctx, lessonID, question); err != nil {
			return 0, err
		}
	}

	return lessonID, nil
}

func (seed *seedTx) ensureEmbeddedQuestion(
	ctx context.Context,
	lessonID int64,
	question lessonQuestionSeed,
) error {
	var questionID int64
	err := seed.tx.QueryRowContext(ctx, `
		SELECT q.id
		FROM lesson_question lq
		JOIN question q
		  ON q.id = lq.question_id
		WHERE lq.lesson_id = ? AND lq.sort_order = ?
		LIMIT 1
	`, lessonID, question.SortOrder).Scan(&questionID)
	if err == nil {
		if _, err := seed.tx.ExecContext(ctx, `
			UPDATE question
			SET title = ?,
			    description = ?,
			    starter_code = ?,
			    reference_code = ?,
			    test_cases = ?
			WHERE id = ?
		`, question.Title, question.Description, question.StarterCode, question.ReferenceCode, question.TestCases, questionID); err != nil {
			return fmt.Errorf("update embedded question sort_order %d: %w", question.SortOrder, err)
		}

		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find embedded question sort_order %d: %w", question.SortOrder, err)
	}

	result, err := seed.tx.ExecContext(ctx, `
		INSERT INTO question (title, description, starter_code, reference_code, test_cases)
		VALUES (?, ?, ?, ?, ?)
	`, question.Title, question.Description, question.StarterCode, question.ReferenceCode, question.TestCases)
	if err != nil {
		return fmt.Errorf("insert embedded question sort_order %d: %w", question.SortOrder, err)
	}

	questionID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read inserted embedded question sort_order %d id: %w", question.SortOrder, err)
	}

	if _, err := seed.tx.ExecContext(ctx, `
		INSERT INTO lesson_question (lesson_id, question_id, sort_order)
		VALUES (?, ?, ?)
	`, lessonID, questionID, question.SortOrder); err != nil {
		return fmt.Errorf("insert embedded lesson question sort_order %d: %w", question.SortOrder, err)
	}

	return nil
}

func (seed *seedTx) ensureDemoEnrollment(ctx context.Context, classroomID, studentID, currentLessonID int64) error {
	if currentLessonID == 0 {
		var classroomCurrentLessonID sql.NullInt64
		if err := seed.tx.QueryRowContext(ctx, `
			SELECT current_lesson_id
			FROM classroom
			WHERE id = ?
		`, classroomID).Scan(&classroomCurrentLessonID); err != nil {
			return fmt.Errorf("load demo classroom current lesson: %w", err)
		}
		if classroomCurrentLessonID.Valid {
			currentLessonID = classroomCurrentLessonID.Int64
		}
	}

	if currentLessonID != 0 {
		if _, err := seed.tx.ExecContext(ctx, `
			UPDATE classroom
			SET current_lesson_id = ?
			WHERE id = ?
		`, currentLessonID, classroomID); err != nil {
			return fmt.Errorf("update demo classroom current lesson: %w", err)
		}
	}

	var enrollmentID int64
	var existingCurrentLessonID sql.NullInt64
	err := seed.tx.QueryRowContext(ctx, `
		SELECT id, current_lesson_id
		FROM enrollment
		WHERE classroom_id = ? AND student_id = ?
		LIMIT 1
	`, classroomID, studentID).Scan(&enrollmentID, &existingCurrentLessonID)
	if err == nil {
		if currentLessonID == 0 {
			return nil
		}

		if existingCurrentLessonID.Valid && existingCurrentLessonID.Int64 == currentLessonID {
			return nil
		}

		if _, err := seed.tx.ExecContext(ctx, `
			UPDATE enrollment
			SET current_lesson_id = ?
			WHERE id = ?
		`, currentLessonID, enrollmentID); err != nil {
			return fmt.Errorf("update demo enrollment current lesson: %w", err)
		}

		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find demo enrollment: %w", err)
	}

	if currentLessonID == 0 {
		if _, err := seed.tx.ExecContext(ctx, `
			INSERT INTO enrollment (classroom_id, student_id)
			VALUES (?, ?)
		`, classroomID, studentID); err != nil {
			return fmt.Errorf("insert demo enrollment: %w", err)
		}

		return nil
	}

	if _, err := seed.tx.ExecContext(ctx, `
		INSERT INTO enrollment (classroom_id, student_id, current_lesson_id)
		VALUES (?, ?, ?)
	`, classroomID, studentID, currentLessonID); err != nil {
		return fmt.Errorf("insert demo enrollment: %w", err)
	}

	return nil
}
