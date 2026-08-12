// Implements question creation, editing, and rules that prevent core judging fields from changing after submissions exist.

package question

import (
	"context"
	"encoding/json"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/logger"
)

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

func (service *service) CreateQuestion(
	ctx context.Context,
	title, description, starterCode, referenceCode string,
	testCases json.RawMessage,
) (Question, error) {
	normalizedTestCases, err := normalizeJSON(testCases, errInvalidTestCases)
	if err != nil {
		return Question{}, errs.Unavailable(err)
	}

	return service.repo.CreateQuestion(
		ctx,
		title,
		description,
		starterCode,
		referenceCode,
		normalizedTestCases,
	)
}

func (service *service) ListQuestions(ctx context.Context) ([]Question, error) {
	return service.repo.ListQuestions(ctx)
}

func (service *service) GetQuestion(ctx context.Context, questionID int64) (Question, error) {
	question, err := service.repo.FindQuestionByID(ctx, questionID)
	if err != nil {
		return Question{}, errs.UnavailableIfNoRows(err)
	}

	return question, nil
}

func (service *service) GetStudentQuestion(
	ctx context.Context,
	classroomID, studentID, lessonQuestionID int64,
) (StudentQuestion, error) {
	if classroomID <= 0 || studentID <= 0 || lessonQuestionID <= 0 {
		return StudentQuestion{}, errs.Unavailable(nil)
	}

	question, err := service.repo.FindStudentQuestionByLessonQuestionID(ctx, classroomID, studentID, lessonQuestionID)
	if err != nil {
		return StudentQuestion{}, errs.UnavailableIfNoRows(err)
	}

	return question, nil
}
