// Implements progress aggregation and statistics for the teacher dashboard.

package progress

import (
	"context"

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

func (service *service) GetClassroomProgress(ctx context.Context, teacherID, classroomID int64) ([]StudentProgress, error) {
	exists, err := service.repo.ClassroomExistsForTeacher(ctx, classroomID, teacherID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.Unavailable(nil)
	}

	return service.repo.ListStudentProgressByClassroom(ctx, classroomID, teacherID)
}

func (service *service) ListClassroomSubmissions(ctx context.Context, teacherID, classroomID int64) ([]Submission, error) {
	exists, err := service.repo.ClassroomExistsForTeacher(ctx, classroomID, teacherID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.Unavailable(nil)
	}

	return service.repo.ListSubmissionsByClassroom(ctx, classroomID, teacherID)
}

func (service *service) GetClassroomSubmission(ctx context.Context, teacherID, classroomID, submissionID int64) (Submission, error) {
	exists, err := service.repo.ClassroomExistsForTeacher(ctx, classroomID, teacherID)
	if err != nil {
		return Submission{}, err
	}
	if !exists {
		return Submission{}, errs.Unavailable(nil)
	}

	item, err := service.repo.FindSubmissionByClassroom(ctx, classroomID, teacherID, submissionID)
	if err != nil {
		return Submission{}, errs.UnavailableIfNoRows(err)
	}

	return item, nil
}

func (service *service) DeleteClassroomSubmission(ctx context.Context, teacherID, classroomID, submissionID int64) error {
	exists, err := service.repo.ClassroomExistsForTeacher(ctx, classroomID, teacherID)
	if err != nil {
		return err
	}
	if !exists {
		return errs.Unavailable(nil)
	}

	return errs.UnavailableIfNoRows(service.repo.DeleteSubmissionByClassroom(ctx, classroomID, teacherID, submissionID))
}
