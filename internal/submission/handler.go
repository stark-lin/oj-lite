// Handles student code submission and self-service submission history APIs.

package submission

import (
	"errors"

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

func (module *Module) CreateSubmission(c *gin.Context) {
	module.handler.CreateSubmission(c)
}

func (module *Module) ListSubmissions(c *gin.Context) {
	module.handler.ListSubmissions(c)
}

func (module *Module) ListQuestionSubmissions(c *gin.Context) {
	module.handler.ListQuestionSubmissions(c)
}

func (module *Module) GetSubmission(c *gin.Context) {
	module.handler.GetSubmission(c)
}

func (handler *handler) CreateSubmission(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	var request createSubmissionRequest
	if !httpx.BindJSONOrNotFound(c, &request) {
		return
	}

	submission, err := handler.service.CreateSubmission(
		c.Request.Context(),
		currentUser.ClassroomID,
		currentUser.ID,
		request.LessonQuestionID,
		request.SourceCode,
	)
	if err != nil {
		switch {
		case errors.Is(err, errLessonQuestionRequired):
			httpx.AbortNotFoundDetails(c, "lesson_question_id is required", gin.H{
				"field": "lesson_question_id",
			})
			return
		case errors.Is(err, errSourceCodeRequired):
			httpx.AbortNotFoundDetails(c, "source_code is required", gin.H{
				"field": "source_code",
			})
			return
		case errors.Is(err, errSourceCodeTooLarge):
			httpx.AbortNotFoundDetails(c, "source_code is too large", gin.H{
				"field":     "source_code",
				"max_bytes": maxSubmissionSourceCodeBytes,
			})
			return
		case errs.IsUnavailable(err):
			httpx.AbortNotFound(c, "question not found")
			return
		default:
			handler.log.Errorf(
				"create submission failed: student_id=%d classroom_id=%d lesson_question_id=%d err=%v",
				currentUser.ID,
				currentUser.ClassroomID,
				request.LessonQuestionID,
				err,
			)
			httpx.AbortInternal(c, err)
			return
		}
	}

	httpx.Created(c, gin.H{
		"submission": newSubmissionDTO(submission, false),
	})
}

func (handler *handler) ListSubmissions(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	items, err := handler.service.ListSubmissions(c.Request.Context(), currentUser.ClassroomID, currentUser.ID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "student not found")
			return
		}

		handler.log.Errorf("list submissions failed: student_id=%d classroom_id=%d err=%v", currentUser.ID, currentUser.ClassroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	dtos := make([]submissionDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, newSubmissionDTO(item, false))
	}

	httpx.OK(c, gin.H{
		"submissions": dtos,
	})
}

func (handler *handler) ListQuestionSubmissions(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	lessonQuestionID, ok := httpx.PathParamInt64OrNotFound(c, "lessonQuestionId")
	if !ok {
		return
	}

	items, err := handler.service.ListQuestionSubmissions(c.Request.Context(), currentUser.ClassroomID, currentUser.ID, lessonQuestionID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "question not found")
			return
		}

		handler.log.Errorf(
			"list question submissions failed: student_id=%d classroom_id=%d lesson_question_id=%d err=%v",
			currentUser.ID,
			currentUser.ClassroomID,
			lessonQuestionID,
			err,
		)
		httpx.AbortInternal(c, err)
		return
	}

	dtos := make([]submissionDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, newSubmissionDTO(item, false))
	}

	httpx.OK(c, gin.H{
		"submissions": dtos,
	})
}

func (handler *handler) GetSubmission(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	submissionID, ok := httpx.PathParamInt64OrNotFound(c, "submissionId")
	if !ok {
		return
	}

	item, err := handler.service.GetSubmission(c.Request.Context(), currentUser.ClassroomID, currentUser.ID, submissionID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "submission not found")
			return
		}

		handler.log.Errorf(
			"get submission failed: student_id=%d classroom_id=%d submission_id=%d err=%v",
			currentUser.ID,
			currentUser.ClassroomID,
			submissionID,
			err,
		)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"submission": newSubmissionDTO(item, true),
	})
}

func (handler *handler) notImplemented(c *gin.Context, action string) {
	handler.log.Warnf("submission scaffold hit: action=%s method=%s path=%s", action, c.Request.Method, c.FullPath())
	httpx.AbortNotImplemented(c, "submission scaffold endpoint is not implemented yet", gin.H{
		"module": "submission",
		"action": action,
		"method": c.Request.Method,
		"path":   c.FullPath(),
	})
}
