package routes

import (
	"github.com/go-chi/chi/v5"
)

func galleryRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// Public Routes
	mux.Get("/", handlerRepo.Gallery.GetAllGallery)

	// Protected Routes (Apply Auth Middleware here if needed)
	mux.Group(func(r chi.Router) {
		// mux.Use(app.Auth) // Uncomment if you have auth middleware

		// Create: POST /gallery
		mux.Post("/", handlerRepo.Gallery.CreateGallery)

		// Update: POST /gallery/{id} (Form Data with ID) or PUT /gallery
		// The JS frontend sends POST to /gallery/{id} with _method=PUT,
		// but the handler reads ID from FormValue("id"), so this works.
		mux.Post("/{id}", handlerRepo.Gallery.UpdateGallery)
		mux.Put("/", handlerRepo.Gallery.UpdateGallery)

		// Delete: DELETE /gallery?id=...
		mux.Delete("/", handlerRepo.Gallery.DeleteGallery)
	})

	return mux
}
