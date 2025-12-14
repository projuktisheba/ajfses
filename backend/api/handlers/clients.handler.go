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

	"github.com/go-chi/chi/v5"
	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"github.com/projuktisheba/ajfses/backend/internal/utils"
)

// ClientHandler is the new handler struct for client operations.
type ClientHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newClientHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) ClientHandler {
	// IMPORTANT: This assumes dbrepo.DBRepository has a field 'ClientRepo'
	// (e.g., DB.ClientRepo.Create, DB.ClientRepo.GetAll, etc.)
	return ClientHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// CreateClient handles the 3-step process: DB Insert -> File Save -> DB Update
func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Multipart Form (10MB limit)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.errorLog.Println("ERROR_CreateClient_01: parsing form:", err)
		utils.BadRequest(w, errors.New("file too large or invalid form data"))
		return
	}

	// 2. Extract Text Data (Fields adapted for the Client model)
	name := strings.TrimSpace(r.FormValue("name"))
	area := strings.TrimSpace(r.FormValue("area"))
	serviceName := strings.TrimSpace(r.FormValue("service_name"))
	serviceDate := strings.TrimSpace(r.FormValue("service_date"))
	status := strings.TrimSpace(r.FormValue("status"))
	note := strings.TrimSpace(r.FormValue("note"))

	if name == "" || area == "" {
		utils.BadRequest(w, errors.New("name and area are required"))
		return
	}

	// 3. STEP ONE: Save data to Database (to generate ID)
	newClient := &models.Client{
		Name:        name,
		Area:        area,
		ServiceName: serviceName,
		ServiceDate: serviceDate,
		Status:      status,
		Note:        note,
		ImageLink:   "", // Empty initially
	}

	id, err := h.DB.ClientRepo.Create(r.Context(), newClient)
	if err != nil {
		h.errorLog.Println("ERROR_CreateClient_03: db create:", err)
		utils.ServerError(w, errors.New("failed to save client info"))
		return
	}

	// 4. STEP TWO: Save Image to File System
	file, header, err := r.FormFile("profileImage")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			// No image uploaded, just return success with the ID
			h.respondSuccess(w, id, "Client created (no image uploaded)")
			return
		}
		h.errorLog.Println("ERROR_CreateClient_04: retrieving file:", err)
		utils.BadRequest(w, errors.New("invalid image file"))
		return
	}
	defer file.Close()

	// Create filename pattern: id_name.ext (e.g., 15_Acme_Corp.jpg)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	safeName := strings.ReplaceAll(name, " ", "_")
	filename := fmt.Sprintf("%d_%s%s", id, safeName, ext)

	// Use a 'clients' specific directory
	storagePath := filepath.Join("data", "images", "clients")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		h.errorLog.Println("ERROR_CreateClient_05: mkdir:", err)
		utils.ServerError(w, errors.New("server storage error"))
		return
	}

	fullPath := filepath.Join(storagePath, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		h.errorLog.Println("ERROR_CreateClient_06: create file:", err)
		utils.ServerError(w, errors.New("failed to create image file"))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.errorLog.Println("ERROR_CreateClient_07: save file:", err)
		utils.ServerError(w, errors.New("failed to save image content"))
		return
	}

	// 5. STEP THREE: Update Database with Image Link
	err = h.DB.ClientRepo.UpdateImageLink(r.Context(), id, filename)
	if err != nil {
		// Log error but do not fail request since data & file are saved
		h.errorLog.Println("ERROR_CreateClient_08: update link:", err)
	}

	h.respondSuccess(w, id, "Client created and image saved successfully")
}

// GetAllClients retrieves a list of all clients, optionally filtered by status.
func (h *ClientHandler) GetAllClients(w http.ResponseWriter, r *http.Request) {
	// Use 'status' query parameter for optional filtering (replacing designation logic)
	status := r.URL.Query().Get("status")

	clients, err := h.DB.ClientRepo.GetAll(r.Context(), status)
	if err != nil {
		h.errorLog.Println("ERROR_GetAllClients_01: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve clients"))
		return
	}
	var response struct {
		Error   bool             `json:"error"`
		Message string           `json:"message"`
		Clients []*models.Client `json:"clients"`
	}
	response.Error = false
	response.Message = "Clients fetched successfully"
	response.Clients = clients
	utils.WriteJSON(w, http.StatusOK, response)
}

// GetClient retrieves a single client by ID.
func (h *ClientHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid client ID"))
		return
	}

	client, err := h.DB.ClientRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_GetClient_01: db error:", err)
		utils.NotFound(w, "client not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, client)
}

// UpdateClient handles updating client details and allows re-uploading the image.
func (h *ClientHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid client ID"))
		return
	}

	// 1. Fetch existing client to preserve data/paths
	existing, err := h.DB.ClientRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateClient_01: fetch error:", err)
		utils.NotFound(w, "client not found")
		return
	}

	// 2. Parse Multipart Form (10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.errorLog.Println("ERROR_UpdateClient_02: parse form:", err)
		utils.BadRequest(w, errors.New("invalid form data or file too large"))
		return
	}

	// 3. Update Text Fields if provided
	name := strings.TrimSpace(r.FormValue("name"))
	if name != "" {
		existing.Name = name
	}

	area := strings.TrimSpace(r.FormValue("area"))
	if area != "" {
		existing.Area = area
	}

	serviceName := strings.TrimSpace(r.FormValue("service_name"))
	if serviceName != "" {
		existing.ServiceName = serviceName
	}
	serviceDate := strings.TrimSpace(r.FormValue("service_date"))
	if serviceDate != "" {
		existing.ServiceDate = serviceDate
	}

	status := strings.TrimSpace(r.FormValue("status"))
	if status != "" {
		existing.Status = status
	}

	note := strings.TrimSpace(r.FormValue("note"))
	if note != "" {
		existing.Note = note
	}

	// --- File Change Tracking Variables ---
	storagePath := filepath.Join("data", "images", "clients")
	oldImageLink := existing.ImageLink // Store original image link
	var backupFilePath string          // Tracks if a successful backup file was created

	// 4. Handle Optional Image Update
	file, header, err := r.FormFile("profileImage")
	if err == nil {
		// New image uploaded
		defer file.Close()

		// 4.1 Backup the Old Image if one exists
		if oldImageLink != "" {
			oldFullPath := filepath.Join(storagePath, oldImageLink)

			// Create backup filename: e.g., "10_Client.jpg" -> "10_Client_backup.jpg"
			oldExt := filepath.Ext(oldImageLink)
			oldBase := strings.TrimSuffix(oldImageLink, oldExt)
			backupFileName := fmt.Sprintf("%s_backup%s", oldBase, oldExt)
			backupFilePath = filepath.Join(storagePath, backupFileName)

			//  STEP 1: Perform Backup (Rename old file to backup file)
			if err := os.Rename(oldFullPath, backupFilePath); err != nil {
				// We proceed if the file doesn't exist, but log an error if rename fails for another reason.
				if !os.IsNotExist(err) {
					h.errorLog.Println("WARNING_UpdateClient_03a: Failed to backup old client image:", err)
				}
				// Clear the path if backup failed/wasn't needed
				backupFilePath = ""
			}
		}

		// 4.2 Save New Image File
		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		safeName := strings.ReplaceAll(existing.Name, " ", "_") // Use potentially new name
		filename := fmt.Sprintf("%d_%s%s", id, safeName, ext)

		// Ensure storage directory exists
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			h.errorLog.Println("ERROR_UpdateClient_03: mkdir:", err)

			//  Restore backup if it exists
			if backupFilePath != "" {
				restorePath := filepath.Join(storagePath, oldImageLink)
				if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
					h.errorLog.Println("CRITICAL_UpdateClient_03b: Failed to restore backup image:", restoreErr)
				}
			}

			utils.ServerError(w, errors.New("storage error"))
			return
		}

		fullPath := filepath.Join(storagePath, filename)
		dst, err := os.Create(fullPath)
		if err != nil {
			h.errorLog.Println("ERROR_UpdateClient_04: create file:", err)

			// Restore backup if it exists
			if backupFilePath != "" {
				restorePath := filepath.Join(storagePath, oldImageLink)
				if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
					h.errorLog.Println("CRITICAL_UpdateClient_04b: Failed to restore backup image:", restoreErr)
				}
			}

			utils.ServerError(w, errors.New("failed to save new image"))
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			h.errorLog.Println("ERROR_UpdateClient_05: copy file:", err)

			// STEP 3 (Part 1): Restore backup if saving new file failed
			if backupFilePath != "" {
				restorePath := filepath.Join(storagePath, oldImageLink)
				// Delete the partially written new file before restore attempt
				os.Remove(fullPath)
				if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
					h.errorLog.Println("CRITICAL_UpdateClient_05a: Failed to restore backup image:", restoreErr)
				}
			}

			utils.ServerError(w, errors.New("failed to write new image"))
			return
		}

		// Update model with the new, successfully saved image link
		existing.ImageLink = filename
	} else if err != http.ErrMissingFile {
		// Handle other errors besides just missing the file
		h.errorLog.Println("ERROR_UpdateClient_05b: form file:", err)
		utils.BadRequest(w, errors.New("failed to process image upload"))
		return
	}

	// 5. Update Database
	err = h.DB.ClientRepo.Update(r.Context(), existing)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateClient_06: db update:", err)

		// STEP 3 (Part 2): Restore original file if DB update failed, AND a backup was made
		if backupFilePath != "" {
			// 1. Delete the newly uploaded file (whose name is now in existing.ImageLink)
			newFullPath := filepath.Join(storagePath, existing.ImageLink)
			if err := os.Remove(newFullPath); err != nil && !os.IsNotExist(err) {
				h.errorLog.Printf("WARNING_UpdateClient_06b: Failed to clean up new image (%s): %v", newFullPath, err)
			}

			// 2. Restore backup file
			restorePath := filepath.Join(storagePath, oldImageLink)
			if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
				h.errorLog.Println("CRITICAL_UpdateClient_06a: Failed to restore backup image after DB failure:", restoreErr)
			}
		}

		utils.ServerError(w, errors.New("failed to update client"))
		return
	}

	// 6. DB Update Successful: Cleanup
	// STEP 4: Remove backup file
	if backupFilePath != "" {
		if err := os.Remove(backupFilePath); err != nil {
			h.errorLog.Println("WARNING_UpdateClient_07: Failed to remove backup image:", err)
		}
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool           `json:"error"`
		Message string         `json:"message"`
		Data    *models.Client `json:"data"`
	}{
		Error:   false,
		Message: "Client updated successfully",
		Data:    existing,
	})
}

// DeleteClient removes the client from the database and its image from the filesystem.
func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid client ID"))
		return
	}

	// Fetch the client info (need ImageLink for file deletion)
	client, err := h.DB.ClientRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.NotFound(w, "Client not found")
		return
	}

	err = h.DB.ClientRepo.Delete(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_DeleteClient_01: db error:", err)
		utils.ServerError(w, errors.New("failed to delete client"))
		return
	}

	// Silently delete the image from the filesystem
	// Changed directory from 'images' to 'clients'
	os.Remove(filepath.Join("data", "images", "clients", client.ImageLink))

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "Client deleted successfully",
	})
}

func (h *ClientHandler) respondSuccess(w http.ResponseWriter, id int64, msg string) {
	utils.WriteJSON(w, http.StatusCreated, struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}{
		Error:   false,
		Message: msg,
		ID:      id,
	})
}
