// Handles teacher-facing question APIs.

package question

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/auth"
	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/logger"
)

var (
	errInvalidDescription = errors.New("invalid description JSON")
	errInvalidTestCases   = errors.New("invalid test_cases JSON")
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

func (module *Module) CreateQuestion(c *gin.Context) {
	module.handler.CreateQuestion(c)
}

func (module *Module) ListQuestions(c *gin.Context) {
	module.handler.ListQuestions(c)
}

func (module *Module) GetQuestion(c *gin.Context) {
	module.handler.GetQuestion(c)
}

func (module *Module) GetStudentQuestion(c *gin.Context) {
	module.handler.GetStudentQuestion(c)
}

func (handler *handler) CreateQuestion(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	var request createQuestionRequest
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

	question, err := handler.service.CreateQuestion(
		c.Request.Context(),
		request.Title,
		request.StarterCode,
		request.ReferenceCode,
		request.Description,
		request.TestCases,
	)
	if err != nil {
		if errors.Is(err, errInvalidDescription) {
			httpx.AbortNotFoundDetails(c, "description must be a valid JSON object", gin.H{
				"field": "description",
			})
			return
		}

		if errors.Is(err, errInvalidTestCases) {
			httpx.AbortNotFoundDetails(c, "test_cases must be valid JSON", gin.H{
				"field": "test_cases",
			})
			return
		}

		handler.log.Errorf("create question failed: teacher_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.Created(c, gin.H{
		"question": newQuestionDTO(question),
	})
}

func (handler *handler) ListQuestions(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	questions, err := handler.service.ListQuestions(c.Request.Context())
	if err != nil {
		handler.log.Errorf("list questions failed: teacher_id=%d err=%v", currentUser.ID, err)
		httpx.AbortInternal(c, err)
		return
	}

	items := make([]questionDTO, 0, len(questions))
	for _, question := range questions {
		items = append(items, newQuestionDTO(question))
	}

	httpx.OK(c, gin.H{
		"questions": items,
	})
}

func (handler *handler) GetQuestion(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	questionID, ok := httpx.PathParamInt64OrNotFound(c, "questionId")
	if !ok {
		return
	}

	question, err := handler.service.GetQuestion(c.Request.Context(), questionID)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "question not found")
			return
		}

		handler.log.Errorf("get question failed: teacher_id=%d question_id=%d err=%v", currentUser.ID, questionID, err)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"question": newQuestionDTO(question),
	})
}

func (handler *handler) GetStudentQuestion(c *gin.Context) {
	currentUser, ok := auth.GetCurrentUser(c)
	if !ok {
		httpx.AbortUnauthorized(c, "missing authenticated user in request context")
		return
	}

	if currentUser.ClassroomID <= 0 {
		httpx.AbortNotFound(c, "question not found")
		return
	}

	lessonQuestionID, ok := httpx.PathParamInt64OrNotFound(c, "lessonQuestionId")
	if !ok {
		return
	}

	question, err := handler.service.GetStudentQuestion(
		c.Request.Context(),
		currentUser.ClassroomID,
		currentUser.ID,
		lessonQuestionID,
	)
	if err != nil {
		if errs.IsUnavailable(err) {
			httpx.AbortNotFound(c, "question not found")
			return
		}

		handler.log.Errorf(
			"get student question failed: student_id=%d classroom_id=%d lesson_question_id=%d err=%v",
			currentUser.ID,
			currentUser.ClassroomID,
			lessonQuestionID,
			err,
		)
		httpx.AbortInternal(c, err)
		return
	}

	httpx.OK(c, gin.H{
		"question": newStudentQuestionDTO(question),
	})
}

func (handler *handler) notImplemented(c *gin.Context, action string) {
	handler.log.Warnf("question scaffold hit: action=%s method=%s path=%s", action, c.Request.Method, c.FullPath())
	httpx.AbortNotImplemented(c, "question scaffold endpoint is not implemented yet", gin.H{
		"module": "question",
		"action": action,
		"method": c.Request.Method,
		"path":   c.FullPath(),
	})
}
