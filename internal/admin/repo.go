// Provides database reads and writes for admin management APIs.

package admin

import (
	"context"
	"database/sql"
	"fmt"

	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/user"
)

const lessonQuestionSortOrderOffset = 1_000_000_000

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

func (repo *repo) CreateTeacher(ctx context.Context, username, passwordHash string) (Teacher, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES (?, ?, 'teacher', 'active')
	`, username, passwordHash)
	if err != nil {
		return Teacher{}, err
	}

	teacherID, err := result.LastInsertId()
	if err != nil {
		return Teacher{}, fmt.Errorf("read inserted teacher id: %w", err)
	}

	return repo.FindTeacherByID(ctx, teacherID)
}

func (repo *repo) ListTeachers(ctx context.Context) ([]Teacher, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, username, password_hash, role, status, created_at
		FROM user_account
		WHERE role = 'teacher'
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teachers := make([]Teacher, 0)
	for rows.Next() {
		teacher, scanErr := scanTeacher(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		teachers = append(teachers, teacher)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teachers, nil
}

func (repo *repo) FindTeacherByID(ctx context.Context, teacherID int64) (Teacher, error) {
	return scanTeacher(repo.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, status, created_at
		FROM user_account
		WHERE id = ? AND role = 'teacher'
		LIMIT 1
	`, teacherID))
}

func (repo *repo) UsernameExistsForOtherUser(ctx context.Context, username string, userID int64) (bool, error) {
	var exists bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_account
			WHERE username = ? AND id != ?
		)
	`, username, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repo *repo) UpdateTeacher(ctx context.Context, teacherID int64, username, status *string) error {
	var (
		result sql.Result
		err    error
	)

	switch {
	case username != nil && status != nil:
		result, err = repo.db.ExecContext(ctx, `
			UPDATE user_account
			SET username = ?, status = ?
			WHERE id = ? AND role = 'teacher'
		`, *username, *status, teacherID)
	case username != nil:
		result, err = repo.db.ExecContext(ctx, `
			UPDATE user_account
			SET username = ?
			WHERE id = ? AND role = 'teacher'
		`, *username, teacherID)
	case status != nil:
		result, err = repo.db.ExecContext(ctx, `
			UPDATE user_account
			SET status = ?
			WHERE id = ? AND role = 'teacher'
		`, *status, teacherID)
	default:
		return nil
	}
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

func (repo *repo) UpdateTeacherPasswordHash(ctx context.Context, teacherID int64, passwordHash string) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE user_account
		SET password_hash = ?
		WHERE id = ? AND role = 'teacher'
	`, passwordHash, teacherID)
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

func (repo *repo) DeleteOrDisableTeacher(ctx context.Context, teacherID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var exists bool
	if err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_account
			WHERE id = ? AND role = 'teacher'
		)
	`, teacherID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		err = sql.ErrNoRows
		return err
	}

	var classroomCount int
	if err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM classroom
		WHERE teacher_id = ?
	`, teacherID).Scan(&classroomCount); err != nil {
		return err
	}

	if classroomCount > 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE user_account
			SET status = 'disabled'
			WHERE id = ? AND role = 'teacher'
		`, teacherID)
	} else {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM user_account
			WHERE id = ? AND role = 'teacher'
		`, teacherID)
	}
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (repo *repo) LessonSortOrderExists(ctx context.Context, sortOrder int, excludeLessonID int64) (bool, error) {
	var exists bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM lesson
			WHERE sort_order = ? AND id != ?
		)
	`, sortOrder, excludeLessonID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repo *repo) CreateLessonWithQuestions(ctx context.Context, input lessonWrite) (Lesson, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Lesson{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO lesson (title, description, sort_order)
		VALUES (?, ?, ?)
	`, input.Title, input.Description, input.SortOrder)
	if err != nil {
		return Lesson{}, err
	}

	lessonID, err := result.LastInsertId()
	if err != nil {
		return Lesson{}, fmt.Errorf("read inserted lesson id: %w", err)
	}

	for _, question := range input.Questions {
		questionID, insertErr := insertQuestionTx(ctx, tx, question)
		if insertErr != nil {
			err = insertErr
			return Lesson{}, err
		}

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO lesson_question (lesson_id, question_id, sort_order)
			VALUES (?, ?, ?)
		`, lessonID, questionID, question.SortOrder); err != nil {
			return Lesson{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return Lesson{}, err
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

	lessons := make([]Lesson, 0)
	for rows.Next() {
		lesson, scanErr := scanLesson(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for index := range lessons {
		questions, listErr := repo.ListLessonQuestions(ctx, lessons[index].ID)
		if listErr != nil {
			return nil, listErr
		}
		lessons[index].Questions = questions
	}

	return lessons, nil
}

func (repo *repo) FindLessonByID(ctx context.Context, lessonID int64) (Lesson, error) {
	lesson, err := scanLesson(repo.db.QueryRowContext(ctx, `
		SELECT id, title, description, sort_order, created_at
		FROM lesson
		WHERE id = ?
		LIMIT 1
	`, lessonID))
	if err != nil {
		return Lesson{}, err
	}

	questions, err := repo.ListLessonQuestions(ctx, lesson.ID)
	if err != nil {
		return Lesson{}, err
	}
	lesson.Questions = questions

	return lesson, nil
}

func (repo *repo) ListLessonQuestions(ctx context.Context, lessonID int64) ([]LessonQuestion, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT lq.id, q.id, q.title, q.description, q.starter_code, q.reference_code, q.test_cases, lq.sort_order, q.created_at
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

	questions := make([]LessonQuestion, 0)
	for rows.Next() {
		question, scanErr := scanLessonQuestion(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		questions = append(questions, question)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func (repo *repo) ReplaceLessonWithQuestions(ctx context.Context, lessonID int64, input lessonWrite) (Lesson, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Lesson{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		UPDATE lesson
		SET title = ?, description = ?, sort_order = ?
		WHERE id = ?
	`, input.Title, input.Description, input.SortOrder, lessonID)
	if err != nil {
		return Lesson{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Lesson{}, err
	}
	if rowsAffected == 0 {
		err = sql.ErrNoRows
		return Lesson{}, err
	}

	existing, err := listExistingLessonQuestionsTx(ctx, tx, lessonID)
	if err != nil {
		return Lesson{}, err
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE lesson_question
		SET sort_order = sort_order + ?
		WHERE lesson_id = ?
	`, lessonQuestionSortOrderOffset, lessonID); err != nil {
		return Lesson{}, err
	}

	seenQuestionIDs := make(map[int64]struct{}, len(input.Questions))
	for _, question := range input.Questions {
		if question.ID > 0 {
			lessonQuestionID, ok := existing[question.ID]
			if !ok {
				err = sql.ErrNoRows
				return Lesson{}, err
			}

			if _, err = tx.ExecContext(ctx, `
				UPDATE question
				SET title = ?, description = ?, starter_code = ?, reference_code = ?, test_cases = ?
				WHERE id = ?
			`, question.Title, question.Description, question.StarterCode, question.ReferenceCode, question.TestCases, question.ID); err != nil {
				return Lesson{}, err
			}

			if _, err = tx.ExecContext(ctx, `
				UPDATE lesson_question
				SET sort_order = ?
				WHERE id = ?
			`, question.SortOrder, lessonQuestionID); err != nil {
				return Lesson{}, err
			}

			seenQuestionIDs[question.ID] = struct{}{}
			continue
		}

		questionID, insertErr := insertQuestionTx(ctx, tx, question)
		if insertErr != nil {
			err = insertErr
			return Lesson{}, err
		}

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO lesson_question (lesson_id, question_id, sort_order)
			VALUES (?, ?, ?)
		`, lessonID, questionID, question.SortOrder); err != nil {
			return Lesson{}, err
		}
		seenQuestionIDs[questionID] = struct{}{}
	}

	for questionID, lessonQuestionID := range existing {
		if _, ok := seenQuestionIDs[questionID]; ok {
			continue
		}

		if _, err = tx.ExecContext(ctx, `
			DELETE FROM lesson_question
			WHERE id = ?
		`, lessonQuestionID); err != nil {
			return Lesson{}, err
		}

		if _, err = tx.ExecContext(ctx, `
			DELETE FROM question
			WHERE id = ?
			  AND NOT EXISTS (
				  SELECT 1
				  FROM lesson_question
				  WHERE question_id = ?
			  )
		`, questionID, questionID); err != nil {
			return Lesson{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return Lesson{}, err
	}

	return repo.FindLessonByID(ctx, lessonID)
}

func (repo *repo) DeleteLesson(ctx context.Context, lessonID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var exists bool
	if err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM lesson
			WHERE id = ?
		)
	`, lessonID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		err = sql.ErrNoRows
		return err
	}

	existing, err := listExistingLessonQuestionsTx(ctx, tx, lessonID)
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM lesson_question
		WHERE lesson_id = ?
	`, lessonID); err != nil {
		return err
	}

	for questionID := range existing {
		if _, err = tx.ExecContext(ctx, `
			DELETE FROM question
			WHERE id = ?
			  AND NOT EXISTS (
				  SELECT 1
				  FROM lesson_question
				  WHERE question_id = ?
			  )
		`, questionID, questionID); err != nil {
			return err
		}
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM lesson
		WHERE id = ?
	`, lessonID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		err = sql.ErrNoRows
		return err
	}

	err = tx.Commit()
	return err
}

func insertQuestionTx(ctx context.Context, tx *sql.Tx, question lessonQuestionWrite) (int64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO question (title, description, starter_code, reference_code, test_cases)
		VALUES (?, ?, ?, ?, ?)
	`, question.Title, question.Description, question.StarterCode, question.ReferenceCode, question.TestCases)
	if err != nil {
		return 0, err
	}

	questionID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted question id: %w", err)
	}

	return questionID, nil
}

func listExistingLessonQuestionsTx(ctx context.Context, tx *sql.Tx, lessonID int64) (map[int64]int64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT question_id, id
		FROM lesson_question
		WHERE lesson_id = ?
	`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[int64]int64)
	for rows.Next() {
		var (
			questionID       int64
			lessonQuestionID int64
		)
		if err := rows.Scan(&questionID, &lessonQuestionID); err != nil {
			return nil, err
		}
		existing[questionID] = lessonQuestionID
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return existing, nil
}

func scanTeacher(scanner interface{ Scan(dest ...any) error }) (Teacher, error) {
	var teacher Teacher
	err := scanner.Scan(
		&teacher.ID,
		&teacher.Username,
		&teacher.PasswordHash,
		&teacher.Role,
		&teacher.Status,
		&teacher.CreatedAt,
	)
	if err != nil {
		return Teacher{}, err
	}
	if teacher.Role == "" {
		teacher.Role = user.RoleTeacher
	}

	return teacher, nil
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
		&question.LessonQuestionID,
		&question.ID,
		&question.Title,
		&question.Description,
		&question.StarterCode,
		&question.ReferenceCode,
		&question.TestCases,
		&question.SortOrder,
		&question.CreatedAt,
	)
	if err != nil {
		return LessonQuestion{}, err
	}

	return question, nil
}
