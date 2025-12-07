package routes

import (
	"github.com/go-chi/chi/v5"
)

func memberRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Member Routes ========
	// Members

	// The handler we created specifically for multipart/form-data
	mux.Post("/", handlerRepo.Member.CreateMember)
	mux.Get("/chairman", handlerRepo.Member.GetChairmanInfo)
	mux.Get("/list", handlerRepo.Member.GetAllMembers)
	mux.Delete("/", handlerRepo.Member.DeleteMember) //	query parament {id}
	// mux.Get("/{id}", handlerRepo.Member.GetMember)
	// mux.Put("/{id}", handlerRepo.Member.UpdateMember)
	return mux
}
