// Registers `/api/*` business routes, including auth, teacher, and student APIs.

package app

import (
	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/middleware"
)

func (app *App) registerAPIRoutes(router gin.IRouter) {
	api := router.Group("/api")

	api.POST("/login", app.authModule.Login)
	api.POST("/logout", app.authModule.Logout)

	authenticated := api.Group("/")
	authenticated.Use(middleware.APIAuth(app.apiSession, app.log))
	authenticated.GET("/me", app.authModule.GetMe)
	authenticated.POST("/me/password", app.authModule.ChangePassword)

	teacher := api.Group("/teacher")
	teacher.Use(middleware.APIAuth(app.apiSession, app.log, "teacher"))
	teacher.POST("/classrooms", app.classroomModule.CreateClassroom)
	teacher.GET("/classrooms", app.classroomModule.ListClassrooms)
	teacher.GET("/classrooms/:classroomId", app.classroomModule.GetClassroom)
	teacher.POST("/classrooms/:classroomId/students", app.classroomModule.CreateStudent)
	teacher.GET("/classrooms/:classroomId/students", app.classroomModule.ListStudents)
	teacher.GET("/classrooms/:classroomId/students/:studentId", app.classroomModule.GetStudent)
	teacher.PATCH("/classrooms/:classroomId/students/:studentId/name", app.classroomModule.RenameStudent)
	teacher.POST("/classrooms/:classroomId/students/:studentId/reset-password", app.classroomModule.ResetStudentPassword)
	teacher.DELETE("/classrooms/:classroomId/students/:studentId", app.classroomModule.RemoveStudent)
	teacher.GET("/lessons", app.lessonModule.ListLessons)
	teacher.GET("/lessons/:lessonId", app.lessonModule.GetLesson)
	teacher.GET("/questions", app.questionModule.ListQuestions)
	teacher.GET("/questions/:questionId", app.questionModule.GetQuestion)
	teacher.GET("/lessons/:lessonId/questions", app.lessonModule.ListQuestions)
	teacher.GET("/classrooms/:classroomId/lessons", app.classroomModule.ListClassroomLessons)
	teacher.POST("/classrooms/:classroomId/current-lesson", app.classroomModule.AdvanceCurrentLesson)
	teacher.GET("/classrooms/:classroomId/progress", app.progressModule.GetClassroomProgress)
	teacher.GET("/classrooms/:classroomId/submissions", app.progressModule.ListClassroomSubmissions)
	teacher.GET("/classrooms/:classroomId/submissions/:submissionId", app.progressModule.GetClassroomSubmission)
	teacher.DELETE("/classrooms/:classroomId/submissions/:submissionId", app.progressModule.DeleteClassroomSubmission)

	student := api.Group("/student")
	student.Use(middleware.APIAuth(app.apiSession, app.log, "student"))
	student.GET("/current-lesson", app.classroomModule.GetCurrentLesson)
	student.GET("/questions/:lessonQuestionId", app.questionModule.GetStudentQuestion)
	student.GET("/questions/:lessonQuestionId/submissions", app.submissionModule.ListQuestionSubmissions)
	student.POST("/submissions", app.submissionModule.CreateSubmission)
	student.GET("/submissions", app.submissionModule.ListSubmissions)
	student.GET("/submissions/:submissionId", app.submissionModule.GetSubmission)
}
