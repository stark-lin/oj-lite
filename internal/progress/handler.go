// Handles teacher APIs for classroom progress, submission lists, and submission details.

package progress

import (
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

func (module *Module) GetClassroomProgress(c *gin.Context) {
	module.handler.GetClassroomProgress(c)
}

func (module *Module) ListClassroomSubmissions(c *gin.Context) {
	module.handler.ListClassroomSubmissions(c)
}

func (module *Module) GetClassroomSubmission(c *gin.Context) {
	module.handler.GetClassroomSubmission(c)
}

func (module *Module) DeleteClassroomSubmission(c *gin.Context) {
	module.handler.DeleteClassroomSubmission(c)
}

func (handler *handler) GetClassroomProgress(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	items, err := handler.service.GetClassroomProgress(c.Request.Context(), currentUser.ID, classroomID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom not found")
			return
		}

		handler.log.Errorf("get classroom progress failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	dtos := make([]studentProgressDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, newStudentProgressDTO(item))
	}

	httpx.OK(c, gin.H{
		"progress": gin.H{
			"students": dtos,
		},
	})
}

func (handler *handler) ListClassroomSubmissions(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}

	items, err := handler.service.ListClassroomSubmissions(c.Request.Context(), currentUser.ID, classroomID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "classroom not found")
			return
		}

		handler.log.Errorf("list classroom submissions failed: teacher_id=%d classroom_id=%d err=%v", currentUser.ID, classroomID, err)
		httpx.AbortInternal(c, err)
		return
	}

	dtos := make([]submissionSummaryDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, newSubmissionSummaryDTO(item, false))
	}

	httpx.OK(c, gin.H{
		"submissions": dtos,
	})
}

func (handler *handler) GetClassroomSubmission(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}
	submissionID, ok := httpx.PathParamInt64OrNotFound(c, "submissionId")
	if !ok {
		return
	}

	item, err := handler.service.GetClassroomSubmission(c.Request.Context(), currentUser.ID, classroomID, submissionID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "submission not found")
			return
		}

		handler.log.Errorf("get classroom submission failed: teacher_id=%d classroom_id=%d submission_id=%d err=%v", currentUser.ID, classroomID, submissionID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"submission": newSubmissionSummaryDTO(item, true),
	})
}

func (handler *handler) DeleteClassroomSubmission(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	classroomID, ok := httpx.PathParamInt64OrNotFound(c, "classroomId")
	if !ok {
		return
	}
	submissionID, ok := httpx.PathParamInt64OrNotFound(c, "submissionId")
	if !ok {
		return
	}

	err := handler.service.DeleteClassroomSubmission(c.Request.Context(), currentUser.ID, classroomID, submissionID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "submission not found")
			return
		}

		handler.log.Errorf("delete classroom submission failed: teacher_id=%d classroom_id=%d submission_id=%d err=%v", currentUser.ID, classroomID, submissionID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"ok": true,
	})
}

func (handler *handler) notImplemented(c *gin.Context, action string) {
	handler.log.Warnf("progress scaffold hit: action=%s method=%s path=%s", action, c.Request.Method, c.FullPath())
	httpx.AbortNotImplemented(c, "progress scaffold endpoint is not implemented yet", gin.H{
		"module": "progress",
		"action": action,
		"method": c.Request.Method,
		"path":   c.FullPath(),
	})
}
