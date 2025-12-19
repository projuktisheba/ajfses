package routes

import (
	"github.com/go-chi/chi/v5"
)

func teamRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Team Routes ========

	// The handler we created specifically for multipart/form-data
	mux.Group(func(r chi.Router) {
		r.Use(authAdmin)
		r.Post("/", handlerRepo.Team.CreateTeam)
	})
	mux.Get("/list", handlerRepo.Team.GetAllTeams)
	mux.Get("/list/details", handlerRepo.Team.GetAllTeamsAndMembers)
	return mux
}
