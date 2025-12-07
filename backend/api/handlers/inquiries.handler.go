package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"github.com/projuktisheba/ajfses/backend/internal/utils"
)

type InquiryHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newInquiryHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) InquiryHandler {
	return InquiryHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// CreateInquiry handles the submission of a new inquiry.
func (h *InquiryHandler) CreateInquiry(w http.ResponseWriter, r *http.Request) {
	var req models.Inquiry
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_CreateInquiry: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// Basic Validation
	req.Name = strings.TrimSpace(req.Name)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Email = strings.TrimSpace(req.Email)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Message = strings.TrimSpace(req.Message)
	fmt.Println("Inq Data: ", req)
	if req.Name == "" || req.Mobile == "" || req.Email == "" || req.Subject == "" || req.Message == "" {
		utils.BadRequest(w, errors.New("All fields are required"))
		return
	}

	// Set default status if empty
	if req.Status == "" {
		req.Status = "NEW"
	}

	id, err := h.DB.InquiryRepo.Create(r.Context(), &req)
	if err != nil {
		h.errorLog.Println("ERROR_02_CreateInquiry: db error:", err)
		utils.ServerError(w, errors.New("failed to submit inquiry"))
		return
	}

	resp := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}{
		Error:   false,
		Message: "Inquiry submitted successfully",
		ID:      id,
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// GetAllInquiries retrieves a list of all inquiries AND status counts (Admin only).
func (h *InquiryHandler) GetAllInquiries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Fetch the List
	inquiries, err := h.DB.InquiryRepo.GetAll(ctx)
	if err != nil {
		h.errorLog.Println("ERROR_01_GetAllInquiries: db error (list):", err)
		utils.ServerError(w, errors.New("failed to retrieve inquiries"))
		return
	}

	// 2. Fetch the Counts
	counts, err := h.DB.InquiryRepo.GetStatusCounts(ctx)
	if err != nil {
		h.errorLog.Println("ERROR_02_GetAllInquiries: db error (counts):", err)
		// Note: Depending on logic, you might not want to fail the whole request
		// if just counts fail, but usually, it's safer to return error.
		utils.ServerError(w, errors.New("failed to retrieve inquiry stats"))
		return
	}

	// 3. Construct the Combined Response
	// We define an anonymous struct here to shape the JSON
	resp := struct {
		Inquiries []models.Inquiry `json:"inquiries"`
		Counts    map[string]int   `json:"counts"`
	}{
		Inquiries: inquiries,
		Counts:    counts,
	}

	// 4. Send JSON
	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetInquiry retrieves a single inquiry by ID.
func (h *InquiryHandler) GetInquiry(w http.ResponseWriter, r *http.Request) {
	// Assuming chi router is used for URL params, user standard approach otherwise
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid inquiry ID"))
		return
	}

	inquiry, err := h.DB.InquiryRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_01_GetInquiry: db error:", err)
		// Distinguish between not found and server error if possible, defaulting to bad request/not found for simplicity
		utils.BadRequest(w, errors.New("inquiry not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, inquiry)
}

// UpdateInquiry handles updating an inquiry's status or details.
func (h *InquiryHandler) UpdateInquiry(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid inquiry ID"))
		return
	}

	// 1. Fetch existing inquiry
	existing, err := h.DB.InquiryRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_01_UpdateInquiry: fetch error:", err)
		utils.BadRequest(w, errors.New("inquiry not found"))
		return
	}

	// 2. Decode update payload
	var req models.Inquiry
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_02_UpdateInquiry: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// 3. Update fields (preserving existing data if payload fields are empty, or overwriting)
	// Strategy: Update whatever is provided in the request
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Mobile != "" {
		existing.Mobile = req.Mobile
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.Subject != "" {
		existing.Subject = req.Subject
	}
	if req.Message != "" {
		existing.Message = req.Message
	}
	if req.Status != "" {
		// Optional: Add validation for allowed status values here
		existing.Status = req.Status
	}

	// 4. Perform Update
	err = h.DB.InquiryRepo.Update(r.Context(), existing)
	if err != nil {
		h.errorLog.Println("ERROR_03_UpdateInquiry: update error:", err)
		utils.ServerError(w, errors.New("failed to update inquiry"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool            `json:"error"`
		Message string          `json:"message"`
		Data    *models.Inquiry `json:"data"`
	}{
		Error:   false,
		Message: "Inquiry updated successfully",
		Data:    existing,
	})
}

// DeleteInquiry removes an inquiry from the database.
func (h *InquiryHandler) DeleteInquiry(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid inquiry ID"))
		return
	}

	err = h.DB.InquiryRepo.Delete(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_01_DeleteInquiry: db error:", err)
		utils.ServerError(w, errors.New("failed to delete inquiry"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "Inquiry deleted successfully",
	})
}
