// Creates the root router and attaches global middleware.

package app

import (
	"github.com/gin-gonic/gin"

	"oj-lite/internal/platform/httpx"
	"oj-lite/internal/platform/middleware"
)

func (app *App) newRouter() *gin.Engine {
	if app.cfg.HTTP.GinMode != "" {
		gin.SetMode(app.cfg.HTTP.GinMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		app.log.Warnf("set trusted proxies failed: %v", err)
	}

	router.Use(
		middleware.RequestID(),
		middleware.Recovery(app.log),
		middleware.Logging(app.log),
	)

	router.GET("/healthz", func(c *gin.Context) {
		httpx.OK(c, gin.H{
			"service": app.cfg.App.Name,
			"env":     app.cfg.App.Env,
			"status":  "ok",
		})
	})

	app.registerPageRoutes(router)
	app.registerAdminRoutes(router)
	app.registerAPIRoutes(router)

	return router
}
