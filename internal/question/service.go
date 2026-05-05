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
	title, starterCode, referenceCode string,
	description, testCases json.RawMessage,
) (Question, error) {
	normalizedDescription, err := normalizeJSONObject(description, errInvalidDescription)
	if err != nil {
		return Question{}, errs.Unavailable(err)
	}

	normalizedTestCases, err := normalizeJSON(testCases, errInvalidTestCases)
	if err != nil {
		return Question{}, errs.Unavailable(err)
	}

	return service.repo.CreateQuestion(
		ctx,
		title,
		normalizedDescription,
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
