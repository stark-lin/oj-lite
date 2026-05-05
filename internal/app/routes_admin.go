// Registers `/admin/*` local admin routes.

package app

import (
	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/middleware"
)

func (app *App) registerAdminRoutes(router gin.IRouter) {
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.AdminAuth())

	adminGroup.POST("/login", app.adminModule.Login)
	adminGroup.POST("/logout", app.adminModule.Logout)

	protected := adminGroup.Group("/")
	protected.POST("/teachers", app.adminModule.CreateTeacher)
	protected.GET("/teachers", app.adminModule.ListTeachers)
	protected.GET("/teachers/:teacherId", app.adminModule.GetTeacher)
	protected.PATCH("/teachers/:teacherId", app.adminModule.UpdateTeacher)
	protected.POST("/teachers/:teacherId/reset-password", app.adminModule.ResetTeacherPassword)
	protected.DELETE("/teachers/:teacherId", app.adminModule.DeleteTeacher)
	protected.POST("/lessons", app.adminModule.CreateLesson)
	protected.GET("/lessons", app.adminModule.ListLessons)
	protected.GET("/lessons/:lessonId", app.adminModule.GetLesson)
	protected.PUT("/lessons/:lessonId", app.adminModule.ReplaceLesson)
	protected.DELETE("/lessons/:lessonId", app.adminModule.DeleteLesson)
}
