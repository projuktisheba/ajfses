package routes

import (
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/projuktisheba/ajfses/backend/api/handlers"
	"github.com/projuktisheba/ajfses/backend/api/middlewares"
	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"github.com/projuktisheba/ajfses/backend/internal/utils"
)

var handlerRepo *handlers.HandlerRepo

func Routes(host, env string, db *dbrepo.DBRepository, jwt models.JWTConfig, infoLogger, errorLogger *log.Logger) http.Handler {
	mux := chi.NewRouter()

	// --- Global middlewares ---
	mux.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://ajfses.pssoft.xyz"},
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Branch-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	mux.Use(middlewares.Logger) // logger

	// --- Static file serving for images ---
	imageDir := filepath.Join(".", "data", "images")
	fs := http.StripPrefix("/api/v1/images/", http.FileServer(http.Dir(imageDir)))
	mux.Handle("/api/v1/images/*", fs)

	// --- Health check ---
	mux.Get("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		ip := "unknown"
		if conn, err := net.Dial("udp", "1.1.1.1:80"); err == nil {
			defer conn.Close()
			ip = conn.LocalAddr().(*net.UDPAddr).IP.String()
		}
		resp := map[string]any{
			"status":    env,
			"server_ip": ip,
		}
		utils.WriteJSON(w, http.StatusOK, resp)
	})

	//get the handler repo
	handlerRepo = handlers.NewHandlerRepo(host, db, jwt, infoLogger, errorLogger)

	// Mount Auth routes
	mux.Mount("/api/v1/auth", authRoutes())

	// =========== Secure Routes ===========

	// Mount services handler routes
	// mux.Mount("/api/v1/ssl", serviceHandlersRoutes())

	// Mount inquiries handler routes
	mux.Mount("/api/v1/inquiry", inquiryRoutes())
	
	// Mount team handler routes
	mux.Mount("/api/v1/team", teamRoutes())

	// Mount member handler routes
	mux.Mount("/api/v1/member", memberRoutes())

	// Mount gallery handler routes
	mux.Mount("/api/v1/gallery", galleryRoutes())

	return mux
}
