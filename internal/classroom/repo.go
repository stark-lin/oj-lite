// Provides data access for classrooms, student ownership, and the unified lesson plan.

package classroom

import (
	"context"
	"database/sql"
	"errors"
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

func (repo *repo) CreateClassroom(ctx context.Context, teacherID int64, name string) (Classroom, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO classroom (teacher_id, current_lesson_id, name)
		VALUES (
			?,
			(
				SELECT id
				FROM lesson
				ORDER BY sort_order ASC, id ASC
				LIMIT 1
			),
			?
		)
	`, teacherID, name)
	if err != nil {
		return Classroom{}, err
	}

	classroomID, err := result.LastInsertId()
	if err != nil {
		return Classroom{}, fmt.Errorf("read inserted classroom id: %w", err)
	}

	return repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID)
}

func (repo *repo) ListClassroomsByTeacherID(ctx context.Context, teacherID int64) ([]Classroom, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, teacher_id, name, created_at
		FROM classroom
		WHERE teacher_id = ?
		ORDER BY id ASC
	`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classrooms []Classroom
	for rows.Next() {
		classroom, err := scanClassroom(rows)
		if err != nil {
			return nil, err
		}
		classrooms = append(classrooms, classroom)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return classrooms, nil
}

func (repo *repo) FindClassroomByIDAndTeacherID(ctx context.Context, classroomID, teacherID int64) (Classroom, error) {
	return scanClassroom(repo.db.QueryRowContext(ctx, `
		SELECT id, teacher_id, name, created_at
		FROM classroom
		WHERE id = ? AND teacher_id = ?
		LIMIT 1
	`, classroomID, teacherID))
}

func (repo *repo) FindDefaultLessonID(ctx context.Context) (sql.NullInt64, error) {
	var lessonID sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM lesson
		ORDER BY sort_order ASC, id ASC
		LIMIT 1
	`).Scan(&lessonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.NullInt64{}, nil
		}
		return sql.NullInt64{}, err
	}

	return lessonID, nil
}

func (repo *repo) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := repo.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_account
			WHERE username = ?
		)
	`, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repo *repo) CreateStudentAndEnrollment(
	ctx context.Context,
	classroomID int64,
	username, passwordHash string,
	currentLessonID *int64,
) (Student, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return Student{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES (?, ?, 'student', 'active')
	`, username, passwordHash)
	if err != nil {
		return Student{}, err
	}

	studentID, err := result.LastInsertId()
	if err != nil {
		return Student{}, fmt.Errorf("read inserted student id: %w", err)
	}

	if currentLessonID == nil {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO enrollment (classroom_id, student_id)
			VALUES (?, ?)
		`, classroomID, studentID)
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO enrollment (classroom_id, student_id, current_lesson_id)
			VALUES (?, ?, ?)
		`, classroomID, studentID, *currentLessonID)
	}
	if err != nil {
		return Student{}, err
	}

	if err = tx.Commit(); err != nil {
		return Student{}, err
	}

	return repo.FindStudentByIDAndClassroomID(ctx, classroomID, studentID)
}

func (repo *repo) ListStudentsByClassroomIDAndTeacherID(ctx context.Context, classroomID, teacherID int64) ([]Student, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.role, u.status, u.created_at, e.current_lesson_id
		FROM classroom c
		JOIN enrollment e
		  ON e.classroom_id = c.id
		JOIN user_account u
		  ON u.id = e.student_id
		WHERE c.id = ? AND c.teacher_id = ?
		ORDER BY u.id ASC
	`, classroomID, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]Student, 0)
	for rows.Next() {
		student, scanErr := scanStudent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (repo *repo) FindStudentByIDAndClassroomID(ctx context.Context, classroomID, studentID int64) (Student, error) {
	return scanStudent(repo.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.role, u.status, u.created_at, e.current_lesson_id
		FROM enrollment e
		JOIN user_account u
		  ON u.id = e.student_id
		WHERE e.classroom_id = ? AND e.student_id = ?
		LIMIT 1
	`, classroomID, studentID))
}

func (repo *repo) FindStudentByIDAndClassroomIDAndTeacherID(
	ctx context.Context,
	classroomID, teacherID, studentID int64,
) (Student, error) {
	return scanStudent(repo.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.role, u.status, u.created_at, e.current_lesson_id
		FROM classroom c
		JOIN enrollment e
		  ON e.classroom_id = c.id
		JOIN user_account u
		  ON u.id = e.student_id
		WHERE c.id = ? AND c.teacher_id = ? AND e.student_id = ?
		LIMIT 1
	`, classroomID, teacherID, studentID))
}

func (repo *repo) UpdateStudentUsername(ctx context.Context, studentID int64, username string) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE user_account
		SET username = ?
		WHERE id = ? AND role = 'student'
	`, username, studentID)
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

func (repo *repo) UpdateStudentPasswordHash(ctx context.Context, studentID int64, passwordHash string) error {
	result, err := repo.db.ExecContext(ctx, `
		UPDATE user_account
		SET password_hash = ?
		WHERE id = ? AND role = 'student'
	`, passwordHash, studentID)
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

func (repo *repo) RemoveStudentFromClassroom(ctx context.Context, classroomID, teacherID, studentID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		DELETE FROM enrollment
		WHERE classroom_id = ?
		  AND student_id = ?
		  AND EXISTS (
			  SELECT 1
			  FROM classroom c
			  WHERE c.id = ? AND c.teacher_id = ?
		  )
	`, classroomID, studentID, classroomID, teacherID)
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

	var remaining int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM enrollment
		WHERE student_id = ?
	`, studentID).Scan(&remaining)
	if err != nil {
		return err
	}

	if remaining == 0 {
		if _, err = tx.ExecContext(ctx, `
			UPDATE user_account
			SET status = 'disabled'
			WHERE id = ? AND role = 'student'
		`, studentID); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *repo) FindClassroomCurrentLessonID(ctx context.Context, classroomID, teacherID int64) (sql.NullInt64, error) {
	var currentLessonID sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT current_lesson_id
		FROM classroom
		WHERE id = ? AND teacher_id = ?
		LIMIT 1
	`, classroomID, teacherID).Scan(&currentLessonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.NullInt64{}, sql.ErrNoRows
		}
		return sql.NullInt64{}, err
	}

	return currentLessonID, nil
}

func (repo *repo) ListClassroomLessonsByTeacherID(ctx context.Context, classroomID, teacherID int64) ([]ClassroomLesson, error) {
	currentLessonID, err := repo.FindClassroomCurrentLessonID(ctx, classroomID, teacherID)
	if err != nil {
		return nil, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT l.id, l.title, l.description, l.sort_order, l.created_at
		FROM classroom c
		CROSS JOIN lesson l
		WHERE c.id = ? AND c.teacher_id = ?
		ORDER BY l.sort_order ASC, l.id ASC
	`, classroomID, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := make([]ClassroomLesson, 0)
	for rows.Next() {
		var lesson ClassroomLesson
		if err := rows.Scan(&lesson.ID, &lesson.Title, &lesson.Description, &lesson.SortOrder, &lesson.CreatedAt); err != nil {
			return nil, err
		}
		lesson.IsCurrent = currentLessonID.Valid && currentLessonID.Int64 == lesson.ID
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessons, nil
}

func (repo *repo) FindLessonByID(ctx context.Context, lessonID int64) (ClassroomLesson, error) {
	var lesson ClassroomLesson
	err := repo.db.QueryRowContext(ctx, `
		SELECT id, title, description, sort_order, created_at
		FROM lesson
		WHERE id = ?
		LIMIT 1
	`, lessonID).Scan(
		&lesson.ID,
		&lesson.Title,
		&lesson.Description,
		&lesson.SortOrder,
		&lesson.CreatedAt,
	)
	if err != nil {
		return ClassroomLesson{}, err
	}

	return lesson, nil
}

func (repo *repo) SetCurrentLessonForClassroom(ctx context.Context, classroomID, teacherID, lessonID int64) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
		UPDATE classroom
		SET current_lesson_id = ?
		WHERE id = ? AND teacher_id = ?
	`, lessonID, classroomID, teacherID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return sql.ErrNoRows
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE enrollment
		SET current_lesson_id = ?
		WHERE classroom_id = ?
	`, lessonID, classroomID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *repo) FindStudentCurrentLesson(ctx context.Context, classroomID, studentID int64) (StudentCurrentLesson, error) {
	return scanStudentCurrentLesson(repo.db.QueryRowContext(ctx, `
		SELECT l.id, l.title, l.description, l.sort_order, l.created_at
		FROM enrollment e
		JOIN classroom c
		  ON c.id = e.classroom_id
		JOIN lesson l
		  ON l.id = c.current_lesson_id
		WHERE e.classroom_id = ? AND e.student_id = ?
		LIMIT 1
	`, classroomID, studentID))
}

func (repo *repo) ListStudentCurrentLessonQuestions(ctx context.Context, lessonID int64) ([]StudentCurrentLessonQuestion, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT lq.id, q.id, q.title, lq.sort_order
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

	var questions []StudentCurrentLessonQuestion
	for rows.Next() {
		question, err := scanStudentCurrentLessonQuestion(rows)
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

func scanClassroom(scanner interface{ Scan(dest ...any) error }) (Classroom, error) {
	var classroom Classroom
	err := scanner.Scan(
		&classroom.ID,
		&classroom.TeacherID,
		&classroom.Name,
		&classroom.CreatedAt,
	)
	if err != nil {
		return Classroom{}, err
	}

	return classroom, nil
}

func scanStudent(scanner interface{ Scan(dest ...any) error }) (Student, error) {
	var student Student
	err := scanner.Scan(
		&student.ID,
		&student.Username,
		&student.Role,
		&student.Status,
		&student.CreatedAt,
		&student.CurrentLessonID,
	)
	if err != nil {
		return Student{}, err
	}

	return student, nil
}

func scanStudentCurrentLesson(scanner interface{ Scan(dest ...any) error }) (StudentCurrentLesson, error) {
	var lesson StudentCurrentLesson
	err := scanner.Scan(
		&lesson.ID,
		&lesson.Title,
		&lesson.Description,
		&lesson.SortOrder,
		&lesson.CreatedAt,
	)
	if err != nil {
		return StudentCurrentLesson{}, err
	}

	return lesson, nil
}

func scanStudentCurrentLessonQuestion(scanner interface{ Scan(dest ...any) error }) (StudentCurrentLessonQuestion, error) {
	var question StudentCurrentLessonQuestion
	err := scanner.Scan(
		&question.LessonQuestionID,
		&question.QuestionID,
		&question.Title,
		&question.SortOrder,
	)
	if err != nil {
		return StudentCurrentLessonQuestion{}, err
	}

	return question, nil
}
