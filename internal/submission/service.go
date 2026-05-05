// Implements student submission persistence, permission checks, and submission queries.

package submission

import (
	"context"
	"errors"
	"strings"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/logger"
)

const maxSubmissionSourceCodeBytes = 65535

var (
	errLessonQuestionRequired = errors.New("lesson question id is required")
	errSourceCodeRequired     = errors.New("source code is required")
	errSourceCodeTooLarge     = errors.New("source code is too large")
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

func (service *service) CreateSubmission(
	ctx context.Context,
	classroomID, studentID, lessonQuestionID int64,
	sourceCode string,
) (Submission, error) {
	if classroomID <= 0 || studentID <= 0 {
		return Submission{}, errs.Unavailable(nil)
	}
	if lessonQuestionID <= 0 {
		return Submission{}, errs.Unavailable(errLessonQuestionRequired)
	}

	if strings.TrimSpace(sourceCode) == "" {
		return Submission{}, errs.Unavailable(errSourceCodeRequired)
	}
	if len([]byte(sourceCode)) > maxSubmissionSourceCodeBytes {
		return Submission{}, errs.Unavailable(errSourceCodeTooLarge)
	}

	target, err := service.repo.FindSubmissionTargetForStudent(ctx, classroomID, studentID, lessonQuestionID)
	if err != nil {
		return Submission{}, errs.UnavailableIfNoRows(err)
	}

	return service.repo.CreateSubmission(ctx, target, sourceCode)
}

func (service *service) ListSubmissions(
	ctx context.Context,
	classroomID, studentID int64,
) ([]Submission, error) {
	if classroomID <= 0 || studentID <= 0 {
		return nil, errs.Unavailable(nil)
	}

	enrollmentID, err := service.repo.FindEnrollmentIDForStudent(ctx, classroomID, studentID)
	if err != nil {
		return nil, errs.UnavailableIfNoRows(err)
	}

	return service.repo.ListSubmissionsByEnrollment(ctx, enrollmentID)
}

func (service *service) ListQuestionSubmissions(
	ctx context.Context,
	classroomID, studentID, lessonQuestionID int64,
) ([]Submission, error) {
	if classroomID <= 0 || studentID <= 0 || lessonQuestionID <= 0 {
		return nil, errs.Unavailable(nil)
	}

	target, err := service.repo.FindSubmissionTargetForStudent(ctx, classroomID, studentID, lessonQuestionID)
	if err != nil {
		return nil, errs.UnavailableIfNoRows(err)
	}

	return service.repo.ListSubmissionsByEnrollmentAndLessonQuestion(ctx, target.EnrollmentID, target.LessonQuestionID)
}

func (service *service) GetSubmission(
	ctx context.Context,
	classroomID, studentID, submissionID int64,
) (Submission, error) {
	if classroomID <= 0 || studentID <= 0 || submissionID <= 0 {
		return Submission{}, errs.Unavailable(nil)
	}

	enrollmentID, err := service.repo.FindEnrollmentIDForStudent(ctx, classroomID, studentID)
	if err != nil {
		return Submission{}, errs.UnavailableIfNoRows(err)
	}

	item, err := service.repo.FindSubmissionByIDAndEnrollment(ctx, enrollmentID, submissionID)
	if err != nil {
		return Submission{}, errs.UnavailableIfNoRows(err)
	}

	return item, nil
}
