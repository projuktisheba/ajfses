package routes

import (
	"github.com/go-chi/chi/v5"
)

func memberRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Member Routes ========
	// Members

	// The handler we created specifically for multipart/form-data

	mux.Get("/messages/chairman", handlerRepo.Member.GetChairmanMessage)
	mux.Get("/messages/ceo", handlerRepo.Member.GetCEOMessage)
	mux.Get("/list", handlerRepo.Member.GetAllMembers)

	mux.Group(func(r chi.Router) {
		r.Use(authAdmin)
		r.Post("/", handlerRepo.Member.CreateMember)
		r.Delete("/", handlerRepo.Member.DeleteMember) //	query parament {id}
		// mux.Get("/{id}", handlerRepo.Member.GetMember)
		r.Put("/", handlerRepo.Member.UpdateMember) // query parameter {id}

	})
	return mux
}
