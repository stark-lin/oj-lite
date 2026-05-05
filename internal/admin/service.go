// Implements local admin business logic, such as managing teachers and lessons.

package admin

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/password"
	"oj-lite/internal/platform/user"
)

var (
	errUsernameAlreadyExists    = errors.New("username already exists")
	errInvalidUsername          = errors.New("invalid username")
	errInvalidTeacherStatus     = errors.New("invalid teacher status")
	errLessonTitleRequired      = errors.New("lesson title is required")
	errLessonSortOrderInvalid   = errors.New("lesson sort_order must be greater than 0")
	errLessonSortOrderExists    = errors.New("lesson sort_order already exists")
	errQuestionSortOrderExists  = errors.New("question sort_order already exists")
	errQuestionSortOrderInvalid = errors.New("question sort_order must be greater than 0")
	errQuestionMissingTitle     = errors.New("question title is required")
	errQuestionDuplicateID      = errors.New("question id is duplicated in request")
	errInvalidDescription       = errors.New("invalid description JSON")
	errInvalidTestCases         = errors.New("invalid test_cases JSON")
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type service struct {
	log  *logger.Logger
	repo *repo
}

func newService(log *logger.Logger, repo *repo) *service {
	return &service{
		log:  log,
		repo: repo,
	}
}

func (service *service) CreateTeacher(ctx context.Context, username, rawPassword string) (Teacher, error) {
	username, err := normalizeUsername(username)
	if err != nil {
		return Teacher{}, errs.Unavailable(err)
	}

	exists, err := service.repo.UsernameExistsForOtherUser(ctx, username, 0)
	if err != nil {
		return Teacher{}, err
	}
	if exists {
		return Teacher{}, errs.Unavailable(errUsernameAlreadyExists)
	}

	passwordHash, err := password.Hash(rawPassword)
	if err != nil {
		if errors.Is(err, password.ErrInvalidLength) {
			return Teacher{}, errs.Unavailable(err)
		}
		return Teacher{}, err
	}

	return service.repo.CreateTeacher(ctx, username, passwordHash)
}

func (service *service) ListTeachers(ctx context.Context) ([]Teacher, error) {
	return service.repo.ListTeachers(ctx)
}

func (service *service) GetTeacher(ctx context.Context, teacherID int64) (Teacher, error) {
	teacher, err := service.repo.FindTeacherByID(ctx, teacherID)
	if err != nil {
		return Teacher{}, errs.UnavailableIfNoRows(err)
	}

	return teacher, nil
}

func (service *service) UpdateTeacher(ctx context.Context, teacherID int64, username, status *string) (Teacher, error) {
	if _, err := service.repo.FindTeacherByID(ctx, teacherID); err != nil {
		return Teacher{}, errs.UnavailableIfNoRows(err)
	}

	var normalizedUsername *string
	if username != nil {
		value, err := normalizeUsername(*username)
		if err != nil {
			return Teacher{}, errs.Unavailable(err)
		}

		exists, err := service.repo.UsernameExistsForOtherUser(ctx, value, teacherID)
		if err != nil {
			return Teacher{}, err
		}
		if exists {
			return Teacher{}, errs.Unavailable(errUsernameAlreadyExists)
		}

		normalizedUsername = &value
	}

	var normalizedStatus *string
	if status != nil {
		value := strings.TrimSpace(*status)
		if value != user.StatusActive && value != user.StatusDisabled {
			return Teacher{}, errs.Unavailable(errInvalidTeacherStatus)
		}
		normalizedStatus = &value
	}

	if err := service.repo.UpdateTeacher(ctx, teacherID, normalizedUsername, normalizedStatus); err != nil {
		return Teacher{}, errs.UnavailableIfNoRows(err)
	}

	return service.repo.FindTeacherByID(ctx, teacherID)
}

func (service *service) ResetTeacherPassword(ctx context.Context, teacherID int64, rawPassword string) error {
	if _, err := service.repo.FindTeacherByID(ctx, teacherID); err != nil {
		return errs.UnavailableIfNoRows(err)
	}

	passwordHash, err := password.Hash(rawPassword)
	if err != nil {
		if errors.Is(err, password.ErrInvalidLength) {
			return errs.Unavailable(err)
		}
		return err
	}

	return errs.UnavailableIfNoRows(service.repo.UpdateTeacherPasswordHash(ctx, teacherID, passwordHash))
}

func (service *service) DeleteTeacher(ctx context.Context, teacherID int64) error {
	return errs.UnavailableIfNoRows(service.repo.DeleteOrDisableTeacher(ctx, teacherID))
}

func (service *service) CreateLesson(ctx context.Context, request lessonRequest) (Lesson, error) {
	input, err := service.normalizeLessonWrite(request)
	if err != nil {
		return Lesson{}, err
	}

	exists, err := service.repo.LessonSortOrderExists(ctx, input.SortOrder, 0)
	if err != nil {
		return Lesson{}, err
	}
	if exists {
		return Lesson{}, errs.Unavailable(errLessonSortOrderExists)
	}

	return service.repo.CreateLessonWithQuestions(ctx, input)
}

func (service *service) ListLessons(ctx context.Context) ([]Lesson, error) {
	return service.repo.ListLessons(ctx)
}

func (service *service) GetLesson(ctx context.Context, lessonID int64) (Lesson, error) {
	lesson, err := service.repo.FindLessonByID(ctx, lessonID)
	if err != nil {
		return Lesson{}, errs.UnavailableIfNoRows(err)
	}

	return lesson, nil
}

func (service *service) ReplaceLesson(ctx context.Context, lessonID int64, request lessonRequest) (Lesson, error) {
	if _, err := service.repo.FindLessonByID(ctx, lessonID); err != nil {
		return Lesson{}, errs.UnavailableIfNoRows(err)
	}

	input, err := service.normalizeLessonWrite(request)
	if err != nil {
		return Lesson{}, err
	}

	exists, err := service.repo.LessonSortOrderExists(ctx, input.SortOrder, lessonID)
	if err != nil {
		return Lesson{}, err
	}
	if exists {
		return Lesson{}, errs.Unavailable(errLessonSortOrderExists)
	}

	lesson, err := service.repo.ReplaceLessonWithQuestions(ctx, lessonID, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Lesson{}, errs.Unavailable(err)
		}
		return Lesson{}, err
	}

	return lesson, nil
}

func (service *service) DeleteLesson(ctx context.Context, lessonID int64) error {
	return errs.UnavailableIfNoRows(service.repo.DeleteLesson(ctx, lessonID))
}

func (service *service) normalizeLessonWrite(request lessonRequest) (lessonWrite, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return lessonWrite{}, errs.Unavailable(errLessonTitleRequired)
	}
	if request.SortOrder <= 0 {
		return lessonWrite{}, errs.Unavailable(errLessonSortOrderInvalid)
	}

	seenSortOrders := make(map[int]struct{}, len(request.Questions))
	seenQuestionIDs := make(map[int64]struct{}, len(request.Questions))
	questions := make([]lessonQuestionWrite, 0, len(request.Questions))
	for _, item := range request.Questions {
		questionTitle := strings.TrimSpace(item.Title)
		if questionTitle == "" {
			return lessonWrite{}, errs.Unavailable(errQuestionMissingTitle)
		}
		if item.SortOrder <= 0 {
			return lessonWrite{}, errs.Unavailable(errQuestionSortOrderInvalid)
		}
		if _, ok := seenSortOrders[item.SortOrder]; ok {
			return lessonWrite{}, errs.Unavailable(errQuestionSortOrderExists)
		}
		seenSortOrders[item.SortOrder] = struct{}{}

		if item.ID > 0 {
			if _, ok := seenQuestionIDs[item.ID]; ok {
				return lessonWrite{}, errs.Unavailable(errQuestionDuplicateID)
			}
			seenQuestionIDs[item.ID] = struct{}{}
		}

		description, err := normalizeJSONObject(item.Description, errInvalidDescription)
		if err != nil {
			return lessonWrite{}, errs.Unavailable(err)
		}

		testCases, err := normalizeJSON(item.TestCases, errInvalidTestCases)
		if err != nil {
			return lessonWrite{}, errs.Unavailable(err)
		}

		questions = append(questions, lessonQuestionWrite{
			ID:            item.ID,
			Title:         questionTitle,
			Description:   description,
			StarterCode:   item.StarterCode,
			ReferenceCode: item.ReferenceCode,
			TestCases:     testCases,
			SortOrder:     item.SortOrder,
		})
	}

	return lessonWrite{
		Title:       title,
		Description: request.Description,
		SortOrder:   request.SortOrder,
		Questions:   questions,
	}, nil
}

func normalizeUsername(value string) (string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 3 || len(value) > 32 {
		return "", errInvalidUsername
	}
	if !usernamePattern.MatchString(value) {
		return "", errInvalidUsername
	}

	return value, nil
}
