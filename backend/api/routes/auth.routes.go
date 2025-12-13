package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/projuktisheba/ajfses/backend/api/middlewares"
)

// Assuming hRepo, cfg, and errorLog are available to this function.
func authRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// Initialize the AuthJWT middleware factory
	authMiddleware := middlewares.AuthJWT(handlerRepo.JWT, handlerRepo.ErrorLog)

	// ======== Public Route ========
	mux.Post("/signin", handlerRepo.Auth.Signin)

	// ======== SECURED ADMIN ROUTES ========
	// Use a group to apply middleware to multiple secure routes efficiently.
	mux.Route("/admin", func(r chi.Router) {

		// 1. Define ALL middleware first using r.Use()
		r.Use(authMiddleware)

		// 2. Define routes after middleware has been applied
		r.Patch("/reset-password", handlerRepo.Auth.UpdatePassword)

		// If you had more admin routes, they go here:
		// r.Get("/users", hRepo.Admin.ListUsers)
	})

	return mux
}
