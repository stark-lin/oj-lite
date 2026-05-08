package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"oj-lite/internal/platform/user"
)

type service struct {
	repo *repo
}

func newService(repo *repo) *service {
	return &service{
		repo: repo,
	}
}

func (service *service) SeedDemoAccounts(ctx context.Context) error {
	return service.repo.withTx(ctx, func(tx *seedTx) error {
		teacherID, err := tx.ensureDemoUser(ctx, demoTeacherUsername, demoTeacherPassword, user.RoleTeacher)
		if err != nil {
			return err
		}

		demoClassroomID, err := tx.ensureDemoClassroom(ctx, teacherID, demoClassroomName)
		if err != nil {
			return err
		}

		studentID, err := tx.ensureDemoUser(ctx, demoStudentUsername, demoStudentPassword, user.RoleStudent)
		if err != nil {
			return err
		}

		if _, err := tx.ensureDemoClassroom(ctx, teacherID, exampleClassroomName); err != nil {
			return err
		}

		lessons, err := loadEmbeddedLessons()
		if err != nil {
			return err
		}

		var firstLessonID int64
		for _, lesson := range lessons {
			lessonID, err := tx.ensureEmbeddedLesson(ctx, lesson)
			if err != nil {
				return err
			}
			if firstLessonID == 0 {
				firstLessonID = lessonID
			}
		}

		return tx.ensureDemoEnrollment(ctx, demoClassroomID, studentID, firstLessonID)
	})
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
