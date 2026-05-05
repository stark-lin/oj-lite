// Implements business logic for creating and editing lessons and organizing question order.

package lesson

import (
	"context"
	"errors"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/logger"
)

var errLessonSortOrderAlreadyExists = errors.New("lesson sort_order already exists")

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

func (service *service) CreateLesson(ctx context.Context, title, description string, sortOrder int) (Lesson, error) {
	exists, err := service.repo.ExistsLessonSortOrder(ctx, sortOrder)
	if err != nil {
		return Lesson{}, err
	}
	if exists {
		return Lesson{}, errs.Unavailable(errLessonSortOrderAlreadyExists)
	}

	return service.repo.CreateLesson(ctx, title, description, sortOrder)
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

func (service *service) ListQuestions(ctx context.Context, lessonID int64) ([]LessonQuestion, error) {
	if _, err := service.repo.FindLessonByID(ctx, lessonID); err != nil {
		return nil, errs.UnavailableIfNoRows(err)
	}

	return service.repo.ListLessonQuestions(ctx, lessonID)
}
