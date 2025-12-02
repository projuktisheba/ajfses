package routes

import (
	"github.com/go-chi/chi/v5"
)

func authRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Auth Routes ========
	mux.Post("/signin", handlerRepo.Auth.Signin)
	// mux.Post("/signup", handlerRepo.Auth.Signup)
	// mux.Post("/refresh-token", handlerRepo.Auth.RefreshToken)
	// mux.Post("/signout", handlerRepo.Auth.Logout)

	return mux
}
