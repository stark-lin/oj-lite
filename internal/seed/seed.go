package seed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"oj-lite/internal/platform/password"
	"oj-lite/internal/platform/user"
)

const (
	demoTeacherUsername  = "teacher"
	demoTeacherPassword  = "teacher"
	demoStudentUsername  = "student"
	demoStudentPassword  = "student"
	demoClassroomName    = "teacher_demo_classroom"
	exampleClassroomName = "example_classroom"
	exampleLessonTitle   = "example_lesson"
	exampleQuestionTitle = "example_question"
	exampleQuestionDesc  = `{"Statement":"Return the sum of two numbers.","Input":"Two integers a and b.","Output":"Their sum."}`
	exampleStarterCode   = "function solution(a, b)\n    return 0\nend"
	exampleReferenceCode = "function solution(a, b)\n    return a + b\nend"
	exampleTestCases     = `[[1,2],[3,4],[10,-3]]`
	exampleLessonOrder   = 100
)

func SeedDemoAccounts(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin demo seed transaction: %w", err)
	}

	teacherID, err := ensureDemoUser(ctx, tx, demoTeacherUsername, demoTeacherPassword, user.RoleTeacher)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	demoClassroomID, err := ensureDemoClassroom(ctx, tx, teacherID, demoClassroomName)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	studentID, err := ensureDemoUser(ctx, tx, demoStudentUsername, demoStudentPassword, user.RoleStudent)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := ensureDemoClassroom(ctx, tx, teacherID, exampleClassroomName); err != nil {
		_ = tx.Rollback()
		return err
	}

	exampleLessonID, err := ensureDemoLesson(
		ctx,
		tx,
		exampleLessonTitle,
		"An example lesson for the seeded classroom.",
		exampleLessonOrder,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	exampleQuestionID, err := ensureDemoQuestion(
		ctx,
		tx,
		exampleQuestionTitle,
		exampleQuestionDesc,
		exampleStarterCode,
		exampleReferenceCode,
		exampleTestCases,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := ensureDemoLessonQuestion(ctx, tx, exampleLessonID, exampleQuestionID, 1); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := ensureDemoEnrollment(ctx, tx, demoClassroomID, studentID, exampleLessonID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit demo seed transaction: %w", err)
	}

	return nil
}

func ensureDemoUser(ctx context.Context, tx *sql.Tx, username, rawPassword, role string) (int64, error) {
	var accountID int64
	var existingRole string
	err := tx.QueryRowContext(ctx, `
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

	result, err := tx.ExecContext(ctx, `
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

func ensureDemoClassroom(ctx context.Context, tx *sql.Tx, teacherID int64, name string) (int64, error) {
	var classroomID int64
	err := tx.QueryRowContext(ctx, `
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

	result, err := tx.ExecContext(ctx, `
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

func ensureDemoLesson(ctx context.Context, tx *sql.Tx, title, description string, sortOrder int) (int64, error) {
	var lessonID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM lesson
		WHERE title = ?
		LIMIT 1
	`, title).Scan(&lessonID)
	if err == nil {
		return lessonID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("find demo lesson: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO lesson (title, description, sort_order)
		VALUES (?, ?, ?)
	`, title, description, sortOrder)
	if err != nil {
		return 0, fmt.Errorf("insert demo lesson: %w", err)
	}

	lessonID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted demo lesson id: %w", err)
	}

	return lessonID, nil
}

func ensureDemoQuestion(
	ctx context.Context,
	tx *sql.Tx,
	title, description, starterCode, referenceCode, testCases string,
) (int64, error) {
	var questionID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM question
		WHERE title = ?
		LIMIT 1
	`, title).Scan(&questionID)
	if err == nil {
		return questionID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("find demo question: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO question (title, description, starter_code, reference_code, test_cases)
		VALUES (?, ?, ?, ?, ?)
	`, title, description, starterCode, referenceCode, testCases)
	if err != nil {
		return 0, fmt.Errorf("insert demo question: %w", err)
	}

	questionID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted demo question id: %w", err)
	}

	return questionID, nil
}

func ensureDemoLessonQuestion(ctx context.Context, tx *sql.Tx, lessonID, questionID int64, sortOrder int) error {
	var lessonQuestionID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM lesson_question
		WHERE lesson_id = ? AND question_id = ?
		LIMIT 1
	`, lessonID, questionID).Scan(&lessonQuestionID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("find demo lesson question: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO lesson_question (lesson_id, question_id, sort_order)
		VALUES (?, ?, ?)
	`, lessonID, questionID, sortOrder); err != nil {
		return fmt.Errorf("insert demo lesson question: %w", err)
	}

	return nil
}

func ensureDemoEnrollment(ctx context.Context, tx *sql.Tx, classroomID, studentID, currentLessonID int64) error {
	if currentLessonID == 0 {
		var classroomCurrentLessonID sql.NullInt64
		if err := tx.QueryRowContext(ctx, `
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
		if _, err := tx.ExecContext(ctx, `
			UPDATE classroom
			SET current_lesson_id = ?
			WHERE id = ?
		`, currentLessonID, classroomID); err != nil {
			return fmt.Errorf("update demo classroom current lesson: %w", err)
		}
	}

	var enrollmentID int64
	var existingCurrentLessonID sql.NullInt64
	err := tx.QueryRowContext(ctx, `
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

		if _, err := tx.ExecContext(ctx, `
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
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO enrollment (classroom_id, student_id)
			VALUES (?, ?)
		`, classroomID, studentID); err != nil {
			return fmt.Errorf("insert demo enrollment: %w", err)
		}

		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO enrollment (classroom_id, student_id, current_lesson_id)
		VALUES (?, ?, ?)
	`, classroomID, studentID, currentLessonID); err != nil {
		return fmt.Errorf("insert demo enrollment: %w", err)
	}

	return nil
}
