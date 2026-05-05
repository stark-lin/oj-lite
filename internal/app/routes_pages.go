// Registers page routes and returns embedded HTML.

package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/middleware"
)

func (app *App) registerPageRoutes(router gin.IRouter) {
	router.GET("/assets/app.css", app.serveAppCSS)
	router.GET("/assets/app.js", app.serveAppJS)
	router.GET("/", app.serveLoginPage)
	router.GET("/admin", middleware.AdminAuth(), app.serveAdminPage)
	router.GET("/teacher", app.serveTeacherPage)
	router.GET("/student", app.serveStudentPage)
}

func (app *App) serveAppCSS(c *gin.Context) {
	app.renderEmbeddedAsset(c, "app.css", "text/css; charset=utf-8")
}

func (app *App) serveAppJS(c *gin.Context) {
	app.renderEmbeddedAsset(c, "app.js", "application/javascript; charset=utf-8")
}

func (app *App) serveLoginPage(c *gin.Context) {
	app.renderEmbeddedHTML(c, "login.html")
}

func (app *App) serveStudentPage(c *gin.Context) {
	app.serveProtectedPage(c, "student", "student.html")
}

func (app *App) serveAdminPage(c *gin.Context) {
	app.renderEmbeddedHTML(c, "admin.html")
}

func (app *App) serveTeacherPage(c *gin.Context) {
	app.serveProtectedPage(c, "teacher", "teacher.html")
}

func (app *App) serveProtectedPage(c *gin.Context, role, page string) {
	claims, err := app.apiSession.ReadAPISessionCookie(c)
	if err != nil {
		app.apiSession.ClearAPISessionCookie(c)
		c.Redirect(http.StatusFound, "/")
		return
	}

	if claims.Role != role {
		c.Redirect(http.StatusFound, "/")
		return
	}

	if app.apiSession.ShouldRefresh(claims) {
		refreshed := app.apiSession.RefreshClaims(claims)
		if err := app.apiSession.SetAPISessionCookie(c, refreshed); err != nil {
			app.log.Errorf("refresh page session cookie failed: user_id=%d err=%v", claims.UserID, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	app.renderEmbeddedHTML(c, page)
}

func (app *App) renderEmbeddedHTML(c *gin.Context, name string) {
	content, err := readEmbeddedHTML(name)
	if err != nil {
		app.log.Errorf("read embedded html failed: name=%s err=%v", name, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

func (app *App) renderEmbeddedAsset(c *gin.Context, name, contentType string) {
	content, err := readEmbeddedAsset(name)
	if err != nil {
		app.log.Errorf("read embedded asset failed: name=%s err=%v", name, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Data(http.StatusOK, contentType, content)
}
