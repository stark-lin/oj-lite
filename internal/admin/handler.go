// Handles admin-facing teacher and lesson management APIs.

package admin

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/password"
)

type handler struct {
	log     *logger.Logger
	service *service
}

func newHandler(log *logger.Logger, service *service) *handler {
	return &handler{
		log:     log,
		service: service,
	}
}

func (module *Module) Login(c *gin.Context) {
	module.handler.Login(c)
}

func (module *Module) Logout(c *gin.Context) {
	module.handler.Logout(c)
}

func (module *Module) CreateTeacher(c *gin.Context) {
	module.handler.CreateTeacher(c)
}

func (module *Module) ListTeachers(c *gin.Context) {
	module.handler.ListTeachers(c)
}

func (module *Module) GetTeacher(c *gin.Context) {
	module.handler.GetTeacher(c)
}

func (module *Module) UpdateTeacher(c *gin.Context) {
	module.handler.UpdateTeacher(c)
}

func (module *Module) ResetTeacherPassword(c *gin.Context) {
	module.handler.ResetTeacherPassword(c)
}

func (module *Module) DeleteTeacher(c *gin.Context) {
	module.handler.DeleteTeacher(c)
}

func (module *Module) CreateLesson(c *gin.Context) {
	module.handler.CreateLesson(c)
}

func (module *Module) ListLessons(c *gin.Context) {
	module.handler.ListLessons(c)
}

func (module *Module) GetLesson(c *gin.Context) {
	module.handler.GetLesson(c)
}

func (module *Module) ReplaceLesson(c *gin.Context) {
	module.handler.ReplaceLesson(c)
}

func (module *Module) DeleteLesson(c *gin.Context) {
	module.handler.DeleteLesson(c)
}

func (handler *handler) Login(c *gin.Context) {
	httpx.OK(c, gin.H{"ok": true})
}

func (handler *handler) Logout(c *gin.Context) {
	httpx.OK(c, gin.H{"ok": true})
}

func (handler *handler) CreateTeacher(c *gin.Context) {
	var request createTeacherRequest
	if !httpx.BindJSON(c, &request) {
		return
	}

	teacher, err := handler.service.CreateTeacher(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidUsername):
			httpx.AbortValidation(c, "username must be 3-32 characters of letters, numbers, or underscore", gin.H{"field": "username"})
			return
		case errors.Is(err, errUsernameAlreadyExists):
			httpx.AbortValidation(c, "username already exists", gin.H{"field": "username"})
			return
		case errors.Is(err, password.ErrInvalidLength):
			httpx.AbortValidation(c, "password length must be between 7 and 128", gin.H{"field": "password"})
			return
		}

		handler.log.Errorf("create teacher failed: err=%v", err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"teacher": newTeacherDTO(teacher),
	})
}

func (handler *handler) ListTeachers(c *gin.Context) {
	teachers, err := handler.service.ListTeachers(c.Request.Context())
	if err != nil {
		handler.log.Errorf("list teachers failed: err=%v", err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]teacherDTO, 0, len(teachers))
	for _, teacher := range teachers {
		items = append(items, newTeacherDTO(teacher))
	}

	httpx.OK(c, gin.H{
		"teachers": items,
	})
}

func (handler *handler) GetTeacher(c *gin.Context) {
	teacherID, ok := httpx.PathParamInt64(c, "teacherId")
	if !ok {
		return
	}

	teacher, err := handler.service.GetTeacher(c.Request.Context(), teacherID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "teacher not found")
			return
		}

		handler.log.Errorf("get teacher failed: teacher_id=%d err=%v", teacherID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"teacher": newTeacherDTO(teacher),
	})
}

func (handler *handler) UpdateTeacher(c *gin.Context) {
	teacherID, ok := httpx.PathParamInt64(c, "teacherId")
	if !ok {
		return
	}

	var request updateTeacherRequest
	if !httpx.BindJSON(c, &request) {
		return
	}

	teacher, err := handler.service.UpdateTeacher(c.Request.Context(), teacherID, request.Username, request.Status)
	if err != nil {
		switch {
		case errs.IsUnavailable(err) && errors.Is(err, sql.ErrNoRows):
			httpx.AbortNotFound(c, "teacher not found")
			return
		case errors.Is(err, errInvalidUsername):
			httpx.AbortValidation(c, "username must be 3-32 characters of letters, numbers, or underscore", gin.H{"field": "username"})
			return
		case errors.Is(err, errUsernameAlreadyExists):
			httpx.AbortValidation(c, "username already exists", gin.H{"field": "username"})
			return
		case errors.Is(err, errInvalidTeacherStatus):
			httpx.AbortValidation(c, "status must be active or disabled", gin.H{"field": "status"})
			return
		case errs.IsUnavailable(err):
			httpx.AbortNotFound(c, "teacher not found")
			return
		}

		handler.log.Errorf("update teacher failed: teacher_id=%d err=%v", teacherID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"teacher": newTeacherDTO(teacher),
	})
}

func (handler *handler) ResetTeacherPassword(c *gin.Context) {
	teacherID, ok := httpx.PathParamInt64(c, "teacherId")
	if !ok {
		return
	}

	var request resetTeacherPasswordRequest
	if !httpx.BindJSON(c, &request) {
		return
	}

	err := handler.service.ResetTeacherPassword(c.Request.Context(), teacherID, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, password.ErrInvalidLength):
			httpx.AbortValidation(c, "password length must be between 7 and 128", gin.H{"field": "password"})
			return
		case errs.IsUnavailable(err):
			httpx.AbortNotFound(c, "teacher not found")
			return
		}

		handler.log.Errorf("reset teacher password failed: teacher_id=%d err=%v", teacherID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{"ok": true})
}

func (handler *handler) DeleteTeacher(c *gin.Context) {
	teacherID, ok := httpx.PathParamInt64(c, "teacherId")
	if !ok {
		return
	}

	err := handler.service.DeleteTeacher(c.Request.Context(), teacherID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "teacher not found")
			return
		}

		handler.log.Errorf("delete teacher failed: teacher_id=%d err=%v", teacherID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{"ok": true})
}

func (handler *handler) CreateLesson(c *gin.Context) {
	var request lessonRequest
	if !httpx.BindJSON(c, &request) {
		return
	}

	lesson, err := handler.service.CreateLesson(c.Request.Context(), request)
	if err != nil {
		if handled := handleLessonError(c, err); handled {
			return
		}

		handler.log.Errorf("create lesson failed: err=%v", err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"lesson": newLessonDTO(lesson),
	})
}

func (handler *handler) ListLessons(c *gin.Context) {
	lessons, err := handler.service.ListLessons(c.Request.Context())
	if err != nil {
		handler.log.Errorf("list lessons failed: err=%v", err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]lessonDTO, 0, len(lessons))
	for _, lesson := range lessons {
		items = append(items, newLessonDTO(lesson))
	}

	httpx.OK(c, gin.H{
		"lessons": items,
	})
}

func (handler *handler) GetLesson(c *gin.Context) {
	lessonID, ok := httpx.PathParamInt64(c, "lessonId")
	if !ok {
		return
	}

	lesson, err := handler.service.GetLesson(c.Request.Context(), lessonID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "lesson not found")
			return
		}

		handler.log.Errorf("get lesson failed: lesson_id=%d err=%v", lessonID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"lesson": newLessonDTO(lesson),
	})
}

func (handler *handler) ReplaceLesson(c *gin.Context) {
	lessonID, ok := httpx.PathParamInt64(c, "lessonId")
	if !ok {
		return
	}

	var request lessonRequest
	if !httpx.BindJSON(c, &request) {
		return
	}

	lesson, err := handler.service.ReplaceLesson(c.Request.Context(), lessonID, request)
	if err != nil {
		if errs.IsUnavailable(err) && errors.Is(err, sql.ErrNoRows) {
			httpx.AbortNotFound(c, "lesson not found")
			return
		}
		if handled := handleLessonError(c, err); handled {
			return
		}
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "lesson not found")
			return
		}

		handler.log.Errorf("replace lesson failed: lesson_id=%d err=%v", lessonID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"lesson": newLessonDTO(lesson),
	})
}

func (handler *handler) DeleteLesson(c *gin.Context) {
	lessonID, ok := httpx.PathParamInt64(c, "lessonId")
	if !ok {
		return
	}

	err := handler.service.DeleteLesson(c.Request.Context(), lessonID)
	if err != nil {
		switch {
		case errs.IsUnavailable(err):
			httpx.AbortNotFound(c, "lesson not found")
			return
		case isForeignKeyError(err):
			httpx.AbortValidation(c, "lesson is still referenced and cannot be deleted", gin.H{"field": "lesson_id"})
			return
		}

		handler.log.Errorf("delete lesson failed: lesson_id=%d err=%v", lessonID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{"ok": true})
}

func handleLessonError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, errLessonTitleRequired):
		httpx.AbortValidation(c, "title is required", gin.H{"field": "title"})
		return true
	case errors.Is(err, errLessonSortOrderInvalid):
		httpx.AbortValidation(c, "sort_order must be greater than 0", gin.H{"field": "sort_order"})
		return true
	case errors.Is(err, errLessonSortOrderExists):
		httpx.AbortValidation(c, "sort_order already exists", gin.H{"field": "sort_order"})
		return true
	case errors.Is(err, errQuestionMissingTitle):
		httpx.AbortValidation(c, "question title is required", gin.H{"field": "questions[].title"})
		return true
	case errors.Is(err, errQuestionSortOrderInvalid):
		httpx.AbortValidation(c, "question sort_order must be greater than 0", gin.H{"field": "questions[].sort_order"})
		return true
	case errors.Is(err, errQuestionSortOrderExists):
		httpx.AbortValidation(c, "question sort_order must be unique within lesson", gin.H{"field": "questions[].sort_order"})
		return true
	case errors.Is(err, errQuestionDuplicateID):
		httpx.AbortValidation(c, "question id must be unique within lesson", gin.H{"field": "questions[].id"})
		return true
	case errors.Is(err, errInvalidDescription):
		httpx.AbortValidation(c, "description must be a valid JSON object", gin.H{"field": "questions[].description"})
		return true
	case errors.Is(err, errInvalidTestCases):
		httpx.AbortValidation(c, "test_cases must be valid JSON", gin.H{"field": "questions[].test_cases"})
		return true
	case isQuestionFrozenError(err):
		httpx.AbortValidation(c, "question core fields are frozen after the first submission", gin.H{"field": "questions"})
		return true
	case isForeignKeyError(err):
		httpx.AbortValidation(c, "lesson data is still referenced and cannot be updated", gin.H{"field": "lesson"})
		return true
	}

	return false
}

func isForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "FOREIGN KEY constraint failed")
}

func isQuestionFrozenError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "question core fields are frozen after the first submission")
}
