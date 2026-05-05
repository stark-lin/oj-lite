// Defines the application object that owns config, database, services, HTTP server, and other core components.

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"oj-lite/internal/admin"
	"oj-lite/internal/classroom"
	"oj-lite/internal/lesson"
	"oj-lite/internal/platform/auth"
	"oj-lite/internal/platform/config"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/session"
	"oj-lite/internal/platform/user"
	"oj-lite/internal/progress"
	"oj-lite/internal/question"
	"oj-lite/internal/submission"
)

type App struct {
	cfg        config.Config
	db         *sql.DB
	log        *logger.Logger
	engine     *gin.Engine
	server     *http.Server
	apiSession *session.Manager

	adminModule      *admin.Module
	authModule       *auth.Module
	classroomModule  *classroom.Module
	lessonModule     *lesson.Module
	questionModule   *question.Module
	progressModule   *progress.Module
	submissionModule *submission.Module
}

func NewApp(cfg config.Config, log *logger.Logger, database *sql.DB, apiSession *session.Manager) *App {
	userStore := user.NewStore(database, log)

	app := &App{
		cfg:              cfg,
		db:               database,
		log:              log,
		apiSession:       apiSession,
		adminModule:      admin.New(database, log),
		authModule:       auth.New(log, apiSession, userStore),
		classroomModule:  classroom.New(database, log),
		lessonModule:     lesson.New(database, log),
		questionModule:   question.New(database, log),
		progressModule:   progress.New(database, log),
		submissionModule: submission.New(database, log),
	}

	app.engine = app.newRouter()
	app.server = &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      app.engine,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return app
}

func (app *App) Config() config.Config {
	return app.cfg
}

func (app *App) Logger() *logger.Logger {
	return app.log
}

func (app *App) DB() *sql.DB {
	return app.db
}

func (app *App) Router() *gin.Engine {
	return app.engine
}

func (app *App) Run() error {
	app.log.Infof("http server listening on %s", app.server.Addr)

	err := app.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (app *App) Shutdown(ctx context.Context) error {
	var serverErr error
	if app.server != nil {
		serverErr = app.server.Shutdown(ctx)
	}

	var dbErr error
	if app.db != nil {
		dbErr = app.db.Close()
	}

	return errors.Join(serverErr, dbErr)
}
