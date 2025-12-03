package routes

import (
	"github.com/go-chi/chi/v5"
)

func inquiryRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Inquiry Routes ========
	mux.Post("/", handlerRepo.Inquiry.CreateInquiry)

	//Query parameter pageLength, pageIndex, status (optional)
	mux.Get("/", handlerRepo.Inquiry.GetAllInquiries)

	//Query parameter {id}
	mux.Patch("/update-status", handlerRepo.Inquiry.UpdateInquiry)
	//Query parameter {id}
	mux.Delete("/", handlerRepo.Inquiry.DeleteInquiry)

	return mux
}
