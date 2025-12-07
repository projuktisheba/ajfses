package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"github.com/projuktisheba/ajfses/backend/internal/utils"
)

type GalleryHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newGalleryHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) GalleryHandler {
	return GalleryHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// CreateGallery handles the upload of multiple images.
// Pattern: DB Insert -> File Save -> DB Update
func (h *GalleryHandler) CreateGallery(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Multipart Form (30MB limit)
	if err := r.ParseMultipartForm(30 << 20); err != nil {
		h.errorLog.Println("ERROR_CreateGallery_01: parsing form:", err)
		utils.BadRequest(w, errors.New("files too large or invalid form data"))
		return
	}

	// 2. Extract Text Data
	title := strings.TrimSpace(r.FormValue("title"))

	// 3. Retrieve Files
	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		utils.BadRequest(w, errors.New("no images uploaded"))
		return
	}

	// Prepare storage path
	storagePath := filepath.Join("data", "images", "gallery")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		h.errorLog.Println("ERROR_CreateGallery_02: mkdir:", err)
		utils.ServerError(w, errors.New("server storage error"))
		return
	}

	countSuccess := 0

	// 4. Iterate over each uploaded file
	for _, header := range files {

		// --- STEP ONE: Save data to Database (to generate ID) ---
		newItem := &models.GalleryItem{
			Title:     title,
			ImageLink: "", // Empty initially
		}

		id, err := h.DB.GalleryRepo.Create(r.Context(), newItem)
		if err != nil {
			h.errorLog.Println("ERROR_CreateGallery_03: db create:", err)
			continue
		}

		// --- STEP TWO: Save Image to File System ---
		file, err := header.Open()
		if err != nil {
			h.errorLog.Println("ERROR_CreateGallery_04: open file:", err)
			continue
		}

		// Create filename pattern: id_title.ext
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}

		// Simple sanitization for filename
		safeTitle := "gallery"
		if title != "" {
			safeTitle = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return '_'
			}, title)
		}

		filename := fmt.Sprintf("%d_%s%s", id, safeTitle, ext)
		fullPath := filepath.Join(storagePath, filename)

		dst, err := os.Create(fullPath)
		if err != nil {
			h.errorLog.Println("ERROR_CreateGallery_05: create file:", err)
			file.Close()
			continue
		}

		if _, err := io.Copy(dst, file); err != nil {
			h.errorLog.Println("ERROR_CreateGallery_06: save file:", err)
			dst.Close()
			file.Close()
			continue
		}

		dst.Close()
		file.Close()

		// --- STEP THREE: Update Database with Image Link ---
		err = h.DB.GalleryRepo.UpdateImageLink(r.Context(), id, filename)
		if err != nil {
			h.errorLog.Println("ERROR_CreateGallery_07: update link:", err)
		} else {
			countSuccess++
		}
	}

	if countSuccess == 0 {
		utils.ServerError(w, errors.New("failed to save any images"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("%d images uploaded successfully", countSuccess),
	})
}

// UpdateGallery handles updating the title or replacing the image of a gallery item.
func (h *GalleryHandler) UpdateGallery(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL or Form (assuming ID is passed in URL query or path)
	// For this example, assuming it's in the form data or query param named "id"
	// Adjust extraction logic based on your router (e.g., chi or mux)
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// Fallback to checking form value if using FormData
		idStr = r.FormValue("id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		utils.BadRequest(w, errors.New("invalid or missing id"))
		return
	}

	// 2. Parse Multipart Form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.BadRequest(w, errors.New("file too large or invalid form data"))
		return
	}

	// 3. Get Existing Item
	item, err := h.DB.GalleryRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateGallery_01: get by id:", err)
		utils.ServerError(w, errors.New("item not found"))
		return
	}

	// 4. Update Text Fields
	newTitle := strings.TrimSpace(r.FormValue("title"))
	if newTitle != "" {
		item.Title = newTitle
		// Update Title in DB
		if err := h.DB.GalleryRepo.Update(r.Context(), item); err != nil {
			h.errorLog.Println("ERROR_UpdateGallery_02: db update:", err)
			utils.ServerError(w, errors.New("failed to update title"))
			return
		}
	}

	// 5. Handle Image Replacement (Optional)
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// A. Create new filename
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}

		safeTitle := "gallery"
		if item.Title != "" {
			safeTitle = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return '_'
			}, item.Title)
		}

		// Add timestamp to ensure uniqueness/cache busting if replacing
		newFilename := fmt.Sprintf("%d_%s_v2%s", item.ID, safeTitle, ext)
		storagePath := filepath.Join("data", "images")
		newFullPath := filepath.Join(storagePath, newFilename)

		// B. Save New File
		dst, err := os.Create(newFullPath)
		if err != nil {
			h.errorLog.Println("ERROR_UpdateGallery_03: create file:", err)
			utils.ServerError(w, errors.New("failed to create new image file"))
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			h.errorLog.Println("ERROR_UpdateGallery_04: save file:", err)
			utils.ServerError(w, errors.New("failed to save new image content"))
			return
		}

		// C. Delete Old File
		if item.ImageLink != "" {
			oldPath := filepath.Join(storagePath, item.ImageLink)
			if err := os.Remove(oldPath); err != nil {
				h.errorLog.Println("WARNING_UpdateGallery_05: failed to delete old file:", err)
				// Non-fatal error
			}
		}

		// D. Update DB Link
		if err := h.DB.GalleryRepo.UpdateImageLink(r.Context(), item.ID, newFilename); err != nil {
			h.errorLog.Println("ERROR_UpdateGallery_06: db update link:", err)
			utils.ServerError(w, errors.New("failed to update image link"))
			return
		}
	} else if !errors.Is(err, http.ErrMissingFile) {
		// Real error occurred during file retrieval
		h.errorLog.Println("ERROR_UpdateGallery_07: retrieving file:", err)
		utils.BadRequest(w, errors.New("error reading file"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Gallery item updated successfully"})
}

// DeleteGallery removes the item from the DB and the file from the disk.
func (h *GalleryHandler) DeleteGallery(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		utils.BadRequest(w, errors.New("invalid or missing id"))
		return
	}

	// 2. Get Item to find filename
	item, err := h.DB.GalleryRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_DeleteGallery_01: item not found:", err)
		utils.ServerError(w, errors.New("item not found or already deleted"))
		return
	}

	// 3. Delete from DB first
	if err := h.DB.GalleryRepo.Delete(r.Context(), id); err != nil {
		h.errorLog.Println("ERROR_DeleteGallery_02: db delete:", err)
		utils.ServerError(w, errors.New("failed to delete database record"))
		return
	}

	// 4. Delete File from Disk
	if item.ImageLink != "" {
		storagePath := filepath.Join("data", "images")
		fullPath := filepath.Join(storagePath, item.ImageLink)

		if err := os.Remove(fullPath); err != nil {
			h.errorLog.Println("WARNING_DeleteGallery_03: failed to delete file:", err)
			// We respond with success because the DB record is gone, which is the primary concern for the API.
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Gallery item deleted successfully"})
}

// GetAllGallery retrieves all items.
func (h *GalleryHandler) GetAllGallery(w http.ResponseWriter, r *http.Request) {
	limit := 0
	limit, _ = strconv.Atoi(r.URL.Query().Get("max_limit"))
	items, err := h.DB.GalleryRepo.GetAll(r.Context(), limit)
	if err != nil {
		h.errorLog.Println("ERROR_GetAllGallery: db query:", err)
		utils.ServerError(w, errors.New("failed to fetch gallery"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, items)
}
