// Handles teacher-facing lesson APIs.

package lesson

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/auth"
	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
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

func (module *Module) CreateLesson(c *gin.Context) {
	module.handler.CreateLesson(c)
}

func (module *Module) ListLessons(c *gin.Context) {
	module.handler.ListLessons(c)
}

func (module *Module) GetLesson(c *gin.Context) {
	module.handler.GetLesson(c)
}

func (module *Module) ListQuestions(c *gin.Context) {
	module.handler.ListQuestions(c)
}

func (handler *handler) CreateLesson(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	var request createLessonRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	request.Title = strings.TrimSpace(request.Title)
	if request.Title == "" {
		httpx.AbortNotFoundDetails(c, "title is required", gin.H{
			"field": "title",
		})
		return
	}

	if request.SortOrder <= 0 {
		httpx.AbortNotFoundDetails(c, "sort_order must be greater than 0", gin.H{
			"field": "sort_order",
		})
		return
	}

	lesson, err := handler.service.CreateLesson(
		c.Request.Context(),
		request.Title,
		request.Description,
		request.SortOrder,
	)
	if err != nil {
		if errors.Is(err, errLessonSortOrderAlreadyExists) {
			httpx.AbortNotFoundDetails(c, "sort_order already exists", gin.H{
				"field": "sort_order",
			})
			return
		}

		handler.log.Errorf("create lesson failed: teacher_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"lesson": newLessonDTO(lesson),
	})
}

func (handler *handler) ListLessons(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	lessons, err := handler.service.ListLessons(c.Request.Context())
	if err != nil {
		handler.log.Errorf("list lessons failed: teacher_id=%d err=%v", currentUser.ID, err)
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
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	lessonID, ok := httpx.PathParamInt64OrNotFound(c, "lessonId")
	if !ok {
		return
	}

	lesson, err := handler.service.GetLesson(c.Request.Context(), lessonID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "lesson not found")
			return
		}

		handler.log.Errorf("get lesson failed: teacher_id=%d lesson_id=%d err=%v", currentUser.ID, lessonID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"lesson": newLessonDTO(lesson),
	})
}

func (handler *handler) ListQuestions(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	lessonID, ok := httpx.PathParamInt64OrNotFound(c, "lessonId")
	if !ok {
		return
	}

	questions, err := handler.service.ListQuestions(c.Request.Context(), lessonID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "lesson not found")
			return
		}

		handler.log.Errorf("list lesson questions failed: teacher_id=%d lesson_id=%d err=%v", currentUser.ID, lessonID, err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]lessonQuestionDTO, 0, len(questions))
	for _, question := range questions {
		items = append(items, newLessonQuestionDTO(question))
	}

	httpx.OK(c, gin.H{
		"questions": items,
	})
}

func (handler *handler) notImplemented(c *gin.Context, action string) {
	handler.log.Warnf("lesson scaffold hit: action=%s method=%s path=%s", action, c.Request.Method, c.FullPath())
	httpx.AbortNotImplemented(c, "lesson scaffold endpoint is not implemented yet", gin.H{
		"module": "lesson",
		"action": action,
		"method": c.Request.Method,
		"path":   c.FullPath(),
	})
}
