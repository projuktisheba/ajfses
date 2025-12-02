package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projuktisheba/ajfses/backend/api/routes"
	"github.com/projuktisheba/ajfses/backend/internal/config"
	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/driver"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

var app *Application

// serve starts the server and listens for requests
func (app *Application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.Port),
		Handler:           app.appRoutes,
		IdleTimeout:       60 * time.Minute, // optional increase
		ReadTimeout:       10 * time.Minute, // allow large uploads
		ReadHeaderTimeout: 2 * time.Minute,
		WriteTimeout:      60 * time.Minute, // allow long deployment
	}

	app.server = srv
	app.infoLog.Printf("Starting HTTP Back end server in %s mode on port %d", app.config.Env, app.config.Port)
	app.infoLog.Println(".....................................")
	return srv.ListenAndServe()
}

// ShutdownServer gracefully shuts down the server
func (app *Application) ShutdownServer() error {
	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.infoLog.Println("Shutting down the server gracefully...")
	// Shutdown the server with the context
	if err := app.server.Shutdown(ctx); err != nil {
		app.errorLog.Printf("Server forced to shutdown: %s", err)
		return err
	}

	app.infoLog.Println("Server exited gracefully")
	return nil
}

// RunServer is the application entry point
func RunServer(ctx context.Context) error {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	var cfg models.Config
	//get environment variables
	cfg, err := config.Load()
	if err != nil {
		errorLog.Println(err)
		return err
	}
	cfg.JWT = models.JWTConfig{
		SecretKey: cfg.JWT.SecretKey,
		Issuer:    cfg.JWT.Issuer,
		Audience:  cfg.JWT.Audience,
		Algorithm: "HS256",
		Expiry:    time.Hour * 24,
	}

	infoLog.Println(cfg)
	// Connection to database
	var dbConn *pgxpool.Pool
	if cfg.Env == "production" {
		dbConn, err = driver.NewPgxPool(cfg.DB.DSN)
	} else {
		//connect to dev database
		dbConn, err = driver.NewPgxPool(cfg.DB.DEVDSN)
	}

	if err != nil {
		errorLog.Println(err)
		return err
	}
	defer dbConn.Close()

	dbRepo := dbrepo.NewDBRepository(dbConn)
	infoLog.Println("Connected to database")

	// create router instance
	routes := routes.Routes(cfg.Host, cfg.Env, dbRepo, cfg.JWT, infoLog, errorLog)
	//Initiate handlers
	app = &Application{
		config:    cfg,
		infoLog:   infoLog,
		errorLog:  errorLog,
		version:   "1.0.0",
		db:        dbRepo,
		appRoutes: routes,
		ctx:       ctx,
	}

	// Run the server in a separate goroutine so we can wait for shutdown signals
	go func() {
		if err := app.serve(); err != nil {
			errorLog.Printf("Error starting server: %s", err)
		}
	}()

	// Channel to listen for OS interrupt signals (e.g., from Ctrl+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Wait for shutdown signal
	<-stop

	// Call ShutdownServer to gracefully shut down the server
	return app.ShutdownServer()
}

// Stop server from outer module
func StopServer() error {
	return app.ShutdownServer()
}
