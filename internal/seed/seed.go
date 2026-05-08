package seed

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

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
	embeddedLessonCount  = 24
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

	lessons, err := loadEmbeddedLessons()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	var firstLessonID int64
	for _, lesson := range lessons {
		lessonID, err := ensureEmbeddedLesson(ctx, tx, lesson)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if firstLessonID == 0 {
			firstLessonID = lessonID
		}
	}

	if err := ensureDemoEnrollment(ctx, tx, demoClassroomID, studentID, firstLessonID); err != nil {
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

type lessonFile struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	SortOrder   int                  `json:"sort_order"`
	Questions   []lessonQuestionFile `json:"questions"`
}

type lessonQuestionFile struct {
	Title         string          `json:"title"`
	Description   json.RawMessage `json:"description"`
	StarterCode   string          `json:"starter_code"`
	ReferenceCode string          `json:"reference_code"`
	TestCases     json.RawMessage `json:"test_cases"`
	SortOrder     int             `json:"sort_order"`
}

type lessonSeed struct {
	Title       string
	Description string
	SortOrder   int
	Questions   []lessonQuestionSeed
}

type lessonQuestionSeed struct {
	Title         string
	Description   string
	StarterCode   string
	ReferenceCode string
	TestCases     string
	SortOrder     int
}

func loadEmbeddedLessons() ([]lessonSeed, error) {
	lessons := make([]lessonSeed, 0, embeddedLessonCount)
	for index := 1; index <= embeddedLessonCount; index++ {
		name := fmt.Sprintf("%d.json", index)
		content, err := readEmbeddedLesson(name)
		if err != nil {
			return nil, fmt.Errorf("read embedded lesson %q: %w", name, err)
		}

		lesson, err := decodeEmbeddedLesson(name, content, index)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func decodeEmbeddedLesson(name string, content []byte, expectedSortOrder int) (lessonSeed, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()

	var raw lessonFile
	if err := decoder.Decode(&raw); err != nil {
		return lessonSeed{}, fmt.Errorf("decode embedded lesson %q: %w", name, err)
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return lessonSeed{}, fmt.Errorf("decode embedded lesson %q trailing content: %w", name, err)
		}
		return lessonSeed{}, fmt.Errorf("decode embedded lesson %q: multiple JSON values", name)
	}

	return normalizeEmbeddedLesson(name, raw, expectedSortOrder)
}

func normalizeEmbeddedLesson(name string, raw lessonFile, expectedSortOrder int) (lessonSeed, error) {
	title := strings.TrimSpace(raw.Title)
	if title == "" {
		return lessonSeed{}, fmt.Errorf("embedded lesson %q title is required", name)
	}
	if raw.SortOrder != expectedSortOrder {
		return lessonSeed{}, fmt.Errorf("embedded lesson %q sort_order = %d, want %d", name, raw.SortOrder, expectedSortOrder)
	}
	if len(raw.Questions) == 0 {
		return lessonSeed{}, fmt.Errorf("embedded lesson %q must contain at least one question", name)
	}

	seenSortOrders := make(map[int]struct{}, len(raw.Questions))
	questions := make([]lessonQuestionSeed, 0, len(raw.Questions))
	for index, item := range raw.Questions {
		questionTitle := strings.TrimSpace(item.Title)
		if questionTitle == "" {
			return lessonSeed{}, fmt.Errorf("embedded lesson %q question %d title is required", name, index+1)
		}
		if item.SortOrder <= 0 {
			return lessonSeed{}, fmt.Errorf("embedded lesson %q question %d sort_order must be positive", name, index+1)
		}
		if _, exists := seenSortOrders[item.SortOrder]; exists {
			return lessonSeed{}, fmt.Errorf("embedded lesson %q question sort_order %d is duplicated", name, item.SortOrder)
		}
		seenSortOrders[item.SortOrder] = struct{}{}

		description, err := compactJSONObject(item.Description)
		if err != nil {
			return lessonSeed{}, fmt.Errorf("embedded lesson %q question %d description: %w", name, index+1, err)
		}

		testCases, err := compactJSON(item.TestCases)
		if err != nil {
			return lessonSeed{}, fmt.Errorf("embedded lesson %q question %d test_cases: %w", name, index+1, err)
		}

		questions = append(questions, lessonQuestionSeed{
			Title:         questionTitle,
			Description:   description,
			StarterCode:   item.StarterCode,
			ReferenceCode: item.ReferenceCode,
			TestCases:     testCases,
			SortOrder:     item.SortOrder,
		})
	}

	return lessonSeed{
		Title:       title,
		Description: raw.Description,
		SortOrder:   raw.SortOrder,
		Questions:   questions,
	}, nil
}

func compactJSON(raw json.RawMessage) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || !json.Valid(raw) {
		return "", fmt.Errorf("must be valid JSON")
	}

	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		return "", fmt.Errorf("must be valid JSON")
	}

	return buffer.String(), nil
}

func compactJSONObject(raw json.RawMessage) (string, error) {
	normalized, err := compactJSON(raw)
	if err != nil {
		return "", err
	}

	var object map[string]any
	if err := json.Unmarshal([]byte(normalized), &object); err != nil || object == nil {
		return "", fmt.Errorf("must be a JSON object")
	}

	return normalized, nil
}

func ensureEmbeddedLesson(ctx context.Context, tx *sql.Tx, lesson lessonSeed) (int64, error) {
	var lessonID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM lesson
		WHERE sort_order = ?
		LIMIT 1
	`, lesson.SortOrder).Scan(&lessonID)
	if err == nil {
		if _, err := tx.ExecContext(ctx, `
			UPDATE lesson
			SET title = ?, description = ?
			WHERE id = ?
		`, lesson.Title, lesson.Description, lessonID); err != nil {
			return 0, fmt.Errorf("update embedded lesson %d: %w", lesson.SortOrder, err)
		}

		for _, question := range lesson.Questions {
			if err := ensureEmbeddedQuestion(ctx, tx, lessonID, question); err != nil {
				return 0, err
			}
		}

		return lessonID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("find embedded lesson %d: %w", lesson.SortOrder, err)
	}

	result, err := tx.ExecContext(ctx, `
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
		if err := ensureEmbeddedQuestion(ctx, tx, lessonID, question); err != nil {
			return 0, err
		}
	}

	return lessonID, nil
}

func ensureEmbeddedQuestion(
	ctx context.Context,
	tx *sql.Tx,
	lessonID int64,
	question lessonQuestionSeed,
) error {
	var questionID int64
	err := tx.QueryRowContext(ctx, `
		SELECT q.id
		FROM lesson_question lq
		JOIN question q
		  ON q.id = lq.question_id
		WHERE lq.lesson_id = ? AND lq.sort_order = ?
		LIMIT 1
	`, lessonID, question.SortOrder).Scan(&questionID)
	if err == nil {
		if _, err := tx.ExecContext(ctx, `
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

	result, err := tx.ExecContext(ctx, `
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

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO lesson_question (lesson_id, question_id, sort_order)
		VALUES (?, ?, ?)
	`, lessonID, questionID, question.SortOrder); err != nil {
		return fmt.Errorf("insert embedded lesson question sort_order %d: %w", question.SortOrder, err)
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
