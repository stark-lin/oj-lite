// Exposes classroom and student management HTTP APIs.

package classroom

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/auth"
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

func (module *Module) CreateClassroom(c *gin.Context) {
	module.handler.CreateClassroom(c)
}

func (module *Module) ListClassrooms(c *gin.Context) {
	module.handler.ListClassrooms(c)
}

func (module *Module) GetClassroom(c *gin.Context) {
	module.handler.GetClassroom(c)
}

func (module *Module) DeleteClassroom(c *gin.Context) {
	module.handler.DeleteClassroom(c)
}

func (module *Module) CreateStudent(c *gin.Context) {
	module.handler.CreateStudent(c)
}

func (module *Module) ListStudents(c *gin.Context) {
	module.handler.ListStudents(c)
}

func (module *Module) GetStudent(c *gin.Context) {
	module.handler.GetStudent(c)
}

func (module *Module) RenameStudent(c *gin.Context) {
	module.handler.RenameStudent(c)
}

func (module *Module) ResetStudentPassword(c *gin.Context) {
	module.handler.ResetStudentPassword(c)
}

func (module *Module) RemoveStudent(c *gin.Context) {
	module.handler.RemoveStudent(c)
}

func (module *Module) ListClassroomLessons(c *gin.Context) {
	module.handler.ListClassroomLessons(c)
}

func (module *Module) AdvanceCurrentLesson(c *gin.Context) {
	module.handler.AdvanceCurrentLesson(c)
}

func (module *Module) GetCurrentLesson(c *gin.Context) {
	module.handler.GetCurrentLesson(c)
}

func (handler *handler) CreateClassroom(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	var request createClassroomRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" {
		httpx.AbortNotFoundDetails(c, "name is required", gin.H{
			"field": "name",
		})
		return
	}

	classroom, err := handler.service.CreateClassroom(c.Request.Context(), currentUser.ID, request.Name)
	if err != nil {
		handler.log.Errorf("create classroom failed: teacher_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"classroom": newClassroomDTO(classroom),
	})
}

func (handler *handler) ListClassrooms(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classrooms, err := handler.service.ListClassrooms(c.Request.Context(), currentUser.ID)
	if err != nil {
		handler.log.Errorf("list classrooms failed: teacher_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]classroomDTO, 0, len(classrooms))
	for _, classroom := range classrooms {
		items = append(items, newClassroomDTO(classroom))
	}

	httpx.OK(c, gin.H{
		"classrooms": items,
	})
}

func (handler *handler) GetClassroom(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	classroom, err := handler.service.GetClassroom(c.Request.Context(), currentUser.ID, classroomID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom not found")
			return
		}

		handler.log.Errorf("get classroom failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"classroom": newClassroomDTO(classroom),
	})
}

func (handler *handler) DeleteClassroom(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	err := handler.service.DeleteClassroom(c.Request.Context(), currentUser.ID, classroomID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom not found")
			return
		}

		handler.log.Errorf("delete classroom failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"ok": true,
	})
}

func (handler *handler) CreateStudent(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	var request createStudentRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	student, err := handler.service.CreateStudent(
		c.Request.Context(),
		currentUser.ID,
		classroomID,
		request.Username,
		request.Password,
	)
	if err != nil {
		switch {
		case errs.IsUnavailable(err) && !errors.Is(err, errInvalidUsername) && !errors.Is(err, errUsernameAlreadyExists) && !errors.Is(err, password.ErrInvalidLength):
			httpx.AbortNotFound(c, "classroom not found")
			return
		case errors.Is(err, errInvalidUsername):
			httpx.AbortNotFoundDetails(c, "username must be 3..32 chars and contain only A-Za-z0-9_", gin.H{"field": "username"})
			return
		case errors.Is(err, errUsernameAlreadyExists):
			httpx.AbortNotFoundDetails(c, "username already exists", gin.H{"field": "username"})
			return
		case errors.Is(err, password.ErrInvalidLength):
			httpx.AbortNotFoundDetails(c, "password length must be between 7 and 128", gin.H{"field": "password"})
			return
		}

		handler.log.Errorf("create student failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"student": newStudentDTO(student),
	})
}

func (handler *handler) ListStudents(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	students, err := handler.service.ListStudents(c.Request.Context(), currentUser.ID, classroomID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom not found")
			return
		}

		handler.log.Errorf("list students failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]studentDTO, 0, len(students))
	for _, student := range students {
		items = append(items, newStudentDTO(student))
	}

	httpx.OK(c, gin.H{
		"students": items,
	})
}

func (handler *handler) GetStudent(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}
	studentID, ok := httpx.PathParamInt64OrNotFound(c, "studentId")
	if !ok {
		return
	}

	student, err := handler.service.GetStudent(c.Request.Context(), currentUser.ID, classroomID, studentID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "student not found")
			return
		}

		handler.log.Errorf("get student failed: teacher_id=%d classroom_id=%d student_id=%d err=%v", currentUser.ID, classroomID, studentID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"student": newStudentDTO(student),
	})
}

func (handler *handler) RenameStudent(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}
	studentID, ok := httpx.PathParamInt64OrNotFound(c, "studentId")
	if !ok {
		return
	}

	var request renameStudentRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	student, err := handler.service.RenameStudent(c.Request.Context(), currentUser.ID, classroomID, studentID, request.Username)
	if err != nil {
		switch {
		case errs.IsUnavailable(err) && !errors.Is(err, errInvalidUsername) && !errors.Is(err, errUsernameAlreadyExists):
			httpx.AbortNotFound(c, "student not found")
			return
		case errors.Is(err, errInvalidUsername):
			httpx.AbortNotFoundDetails(c, "username must be 3..32 chars and contain only A-Za-z0-9_", gin.H{"field": "username"})
			return
		case errors.Is(err, errUsernameAlreadyExists):
			httpx.AbortNotFoundDetails(c, "username already exists", gin.H{"field": "username"})
			return
		}

		handler.log.Errorf("rename student failed: teacher_id=%d classroom_id=%d student_id=%d err=%v", currentUser.ID, classroomID, studentID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"student": newStudentDTO(student),
	})
}

func (handler *handler) ResetStudentPassword(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}
	studentID, ok := httpx.PathParamInt64OrNotFound(c, "studentId")
	if !ok {
		return
	}

	var request resetStudentPasswordRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	err := handler.service.ResetStudentPassword(c.Request.Context(), currentUser.ID, classroomID, studentID, request.Password)
	if err != nil {
		switch {
		case errs.IsUnavailable(err) && !errors.Is(err, password.ErrInvalidLength):
			httpx.AbortNotFound(c, "student not found")
			return
		case errors.Is(err, password.ErrInvalidLength):
			httpx.AbortNotFoundDetails(c, "password length must be between 7 and 128", gin.H{"field": "password"})
			return
		}

		handler.log.Errorf("reset student password failed: teacher_id=%d classroom_id=%d student_id=%d err=%v", currentUser.ID, classroomID, studentID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"ok": true,
	})
}

func (handler *handler) RemoveStudent(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}
	studentID, ok := httpx.PathParamInt64OrNotFound(c, "studentId")
	if !ok {
		return
	}

	err := handler.service.RemoveStudent(c.Request.Context(), currentUser.ID, classroomID, studentID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "student not found")
			return
		}

		handler.log.Errorf("remove student failed: teacher_id=%d classroom_id=%d student_id=%d err=%v", currentUser.ID, classroomID, studentID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"ok": true,
	})
}

func (handler *handler) ListClassroomLessons(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	lessons, err := handler.service.ListClassroomLessons(c.Request.Context(), currentUser.ID, classroomID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom not found")
			return
		}

		handler.log.Errorf("list classroom lessons failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]classroomLessonDTO, 0, len(lessons))
	for _, lesson := range lessons {
		items = append(items, newClassroomLessonDTO(lesson))
	}

	httpx.OK(c, gin.H{
		"lessons": items,
	})
}

func (handler *handler) AdvanceCurrentLesson(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	var request setCurrentLessonRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}
	if request.LessonID <= 0 {
		httpx.AbortNotFoundDetails(c, "lesson_id must be greater than 0", gin.H{"field": "lesson_id"})
		return
	}

	lesson, err := handler.service.SetCurrentLesson(c.Request.Context(), currentUser.ID, classroomID, request.LessonID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom or lesson not found")
			return
		}

		handler.log.Errorf("set current lesson failed: teacher_id=%d classroom_id=%d lesson_id=%d err=%v", currentUser.ID, classroomID, request.LessonID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"current_lesson": newClassroomLessonDTO(lesson),
	})
}

func (handler *handler) GetCurrentLesson(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	if currentUser.ClassroomID <= 0 {
		httpx.AbortNotFound(c, "current lesson not found")
		return
	}

	lesson, err := handler.service.GetStudentCurrentLesson(c.Request.Context(), currentUser.ClassroomID, currentUser.ID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "current lesson not found")
			return
		}

		handler.log.Errorf(
			"get current lesson failed: student_id=%d classroom_id=%d err=%v",
			currentUser.ID,
			currentUser.ClassroomID,
			err,
		)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"lesson": newStudentCurrentLessonDTO(lesson),
	})
}

func (handler *handler) notImplemented(c *gin.Context, action string) {
	handler.log.Warnf("classroom scaffold hit: action=%s method=%s path=%s", action, c.Request.Method, c.FullPath())
	httpx.AbortNotImplemented(c, "classroom scaffold endpoint is not implemented yet", gin.H{
		"module": "classroom",
		"action": action,
		"method": c.Request.Method,
		"path":   c.FullPath(),
	})
}
