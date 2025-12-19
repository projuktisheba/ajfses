package routes

import (
	"github.com/go-chi/chi/v5"
)

func inquiryRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Inquiry Routes ========
	mux.Post("/", handlerRepo.Inquiry.CreateInquiry)

	mux.Group(func(r chi.Router) {
		r.Use(authAdmin)
		//Query parameter pageLength, pageIndex, status (optional)
		r.Get("/", handlerRepo.Inquiry.GetAllInquiries)

		//Query parameter {id}
		r.Patch("/update-status", handlerRepo.Inquiry.UpdateInquiry)
		//Query parameter {id}
		r.Delete("/", handlerRepo.Inquiry.DeleteInquiry)
	})

	return mux
}
