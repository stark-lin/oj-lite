// Implements classroom creation, student enrollment, unified lesson plans, and current-lesson advancement.

package classroom

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/password"
)

var (
	errUsernameAlreadyExists = errors.New("username already exists")
	errInvalidUsername       = errors.New("invalid username")
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

func (service *service) CreateClassroom(ctx context.Context, teacherID int64, name string) (Classroom, error) {
	return service.repo.CreateClassroom(ctx, teacherID, name)
}

func (service *service) ListClassrooms(ctx context.Context, teacherID int64) ([]Classroom, error) {
	return service.repo.ListClassroomsByTeacherID(ctx, teacherID)
}

func (service *service) GetClassroom(ctx context.Context, teacherID, classroomID int64) (Classroom, error) {
	classroom, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID)
	if err != nil {
		return Classroom{}, errs.UnavailableIfNoRows(err)
	}

	return classroom, nil
}

func (service *service) CreateStudent(
	ctx context.Context,
	teacherID, classroomID int64,
	username, rawPassword string,
) (Student, error) {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}

	username = strings.TrimSpace(username)
	if err := validateUsername(username); err != nil {
		return Student{}, errs.Unavailable(err)
	}

	exists, err := service.repo.UsernameExists(ctx, username)
	if err != nil {
		return Student{}, err
	}
	if exists {
		return Student{}, errs.Unavailable(errUsernameAlreadyExists)
	}

	passwordHash, err := password.Hash(rawPassword)
	if err != nil {
		if errors.Is(err, password.ErrInvalidLength) {
			return Student{}, errs.Unavailable(err)
		}
		return Student{}, err
	}

	currentLessonID, err := service.repo.FindClassroomCurrentLessonID(ctx, classroomID, teacherID)
	if err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}
	if !currentLessonID.Valid {
		defaultLessonID, err := service.repo.FindDefaultLessonID(ctx)
		if err != nil {
			return Student{}, err
		}
		if defaultLessonID.Valid {
			if err := service.repo.SetCurrentLessonForClassroom(ctx, classroomID, teacherID, defaultLessonID.Int64); err != nil {
				return Student{}, errs.UnavailableIfNoRows(err)
			}
			currentLessonID = defaultLessonID
		}
	}

	var nextCurrentLessonID *int64
	if currentLessonID.Valid {
		nextCurrentLessonID = &currentLessonID.Int64
	}

	return service.repo.CreateStudentAndEnrollment(ctx, classroomID, username, passwordHash, nextCurrentLessonID)
}

func (service *service) ListStudents(ctx context.Context, teacherID, classroomID int64) ([]Student, error) {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return nil, errs.UnavailableIfNoRows(err)
	}
	return service.repo.ListStudentsByClassroomIDAndTeacherID(ctx, classroomID, teacherID)
}

func (service *service) GetStudent(ctx context.Context, teacherID, classroomID, studentID int64) (Student, error) {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}

	student, err := service.repo.FindStudentByIDAndClassroomIDAndTeacherID(ctx, classroomID, teacherID, studentID)
	if err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}

	return student, nil
}

func (service *service) RenameStudent(ctx context.Context, teacherID, classroomID, studentID int64, username string) (Student, error) {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}
	if _, err := service.repo.FindStudentByIDAndClassroomIDAndTeacherID(ctx, classroomID, teacherID, studentID); err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}

	username = strings.TrimSpace(username)
	if err := validateUsername(username); err != nil {
		return Student{}, errs.Unavailable(err)
	}

	existing, err := service.repo.UsernameExists(ctx, username)
	if err != nil {
		return Student{}, err
	}
	if existing {
		current, err := service.repo.FindStudentByIDAndClassroomIDAndTeacherID(ctx, classroomID, teacherID, studentID)
		if err != nil {
			return Student{}, errs.UnavailableIfNoRows(err)
		}
		if current.Username != username {
			return Student{}, errs.Unavailable(errUsernameAlreadyExists)
		}
		return current, nil
	}

	if err := service.repo.UpdateStudentUsername(ctx, studentID, username); err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}

	student, err := service.repo.FindStudentByIDAndClassroomIDAndTeacherID(ctx, classroomID, teacherID, studentID)
	if err != nil {
		return Student{}, errs.UnavailableIfNoRows(err)
	}

	return student, nil
}

func (service *service) ResetStudentPassword(
	ctx context.Context,
	teacherID, classroomID, studentID int64,
	rawPassword string,
) error {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return errs.UnavailableIfNoRows(err)
	}
	if _, err := service.repo.FindStudentByIDAndClassroomIDAndTeacherID(ctx, classroomID, teacherID, studentID); err != nil {
		return errs.UnavailableIfNoRows(err)
	}

	passwordHash, err := password.Hash(rawPassword)
	if err != nil {
		if errors.Is(err, password.ErrInvalidLength) {
			return errs.Unavailable(err)
		}
		return err
	}

	return errs.UnavailableIfNoRows(service.repo.UpdateStudentPasswordHash(ctx, studentID, passwordHash))
}

func (service *service) RemoveStudent(ctx context.Context, teacherID, classroomID, studentID int64) error {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return errs.UnavailableIfNoRows(err)
	}
	if _, err := service.repo.FindStudentByIDAndClassroomIDAndTeacherID(ctx, classroomID, teacherID, studentID); err != nil {
		return errs.UnavailableIfNoRows(err)
	}

	return errs.UnavailableIfNoRows(service.repo.RemoveStudentFromClassroom(ctx, classroomID, teacherID, studentID))
}

func (service *service) ListClassroomLessons(ctx context.Context, teacherID, classroomID int64) ([]ClassroomLesson, error) {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return nil, errs.UnavailableIfNoRows(err)
	}
	return service.repo.ListClassroomLessonsByTeacherID(ctx, classroomID, teacherID)
}

func (service *service) SetCurrentLesson(ctx context.Context, teacherID, classroomID, lessonID int64) (ClassroomLesson, error) {
	if _, err := service.repo.FindClassroomByIDAndTeacherID(ctx, classroomID, teacherID); err != nil {
		return ClassroomLesson{}, errs.UnavailableIfNoRows(err)
	}

	lesson, err := service.repo.FindLessonByID(ctx, lessonID)
	if err != nil {
		return ClassroomLesson{}, errs.UnavailableIfNoRows(err)
	}
	lesson.IsCurrent = true

	if err := service.repo.SetCurrentLessonForClassroom(ctx, classroomID, teacherID, lessonID); err != nil {
		return ClassroomLesson{}, errs.UnavailableIfNoRows(err)
	}

	return lesson, nil
}

func (service *service) GetStudentCurrentLesson(ctx context.Context, classroomID, studentID int64) (StudentCurrentLesson, error) {
	lesson, err := service.repo.FindStudentCurrentLesson(ctx, classroomID, studentID)
	if err != nil {
		return StudentCurrentLesson{}, errs.UnavailableIfNoRows(err)
	}

	questions, err := service.repo.ListStudentCurrentLessonQuestions(ctx, lesson.ID)
	if err != nil {
		return StudentCurrentLesson{}, err
	}

	lesson.Questions = questions
	return lesson, nil
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return errInvalidUsername
	}
	if !usernamePattern.MatchString(username) {
		return errInvalidUsername
	}

	return nil
}
