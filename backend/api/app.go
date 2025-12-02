package api

import (
	"context"
	"log"
	"net/http"

	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

// application is the receiver for the various parts of the application
type Application struct {
	config    models.Config
	infoLog   *log.Logger
	errorLog  *log.Logger
	version   string
	appRoutes http.Handler
	db        *dbrepo.DBRepository
	server    *http.Server
	ctx       context.Context
}

// =======================
// Getter Methods
// =======================

// GetHandlers returns the Handlers for routes
func (app *Application) GetHandlers() http.Handler {
	return app.appRoutes
}

// Config returns the application config
func (app *Application) GetConfig() models.Config {
	return app.config
}

// InfoLog returns the info logger
func (app *Application) GetInfoLog() *log.Logger {
	return app.infoLog
}

// ErrorLog returns the error logger
func (app *Application) GetErrorLog() *log.Logger {
	return app.errorLog
}

// AppLogger returns the info logger & error logger
func (app *Application) GetAppLogger() (*log.Logger, *log.Logger) {
	return app.infoLog, app.errorLog
}

// Version returns the app version
func (app *Application) GetVersion() string {
	return app.version
}

// DBRepo returns the database repository
func (app *Application) GetDBRepo() *dbrepo.DBRepository {
	return app.db
}

// Context returns the application context
func (app *Application) Context() context.Context {
	return app.ctx
}

// ServerInstance returns the HTTP server instance
func (app *Application) ServerInstance() *http.Server {
	return app.server
}
