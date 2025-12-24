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

type MemberHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newMemberHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) MemberHandler {
	return MemberHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// CreateMember handles the 3-step process: DB Insert -> File Save -> DB Update
func (h *MemberHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Multipart Form (10MB limit)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.errorLog.Println("ERROR_CreateMember_01: parsing form:", err)
		utils.BadRequest(w, errors.New("file too large or invalid form data"))
		return
	}

	// 2. Extract Text Data
	name := strings.TrimSpace(r.FormValue("name"))
	team := strings.TrimSpace(r.FormValue("team"))
	designation := strings.TrimSpace(r.FormValue("designation"))
	contact := strings.TrimSpace(r.FormValue("contact"))
	note := strings.TrimSpace(r.FormValue("note"))
	showOnHomepage := false
	if strings.TrimSpace(r.FormValue("showOnHome")) == "1" {
		showOnHomepage = true
	}

	if name == "" || team == "" {
		utils.BadRequest(w, errors.New("name and team are required"))
		return
	}

	teamID, err := strconv.Atoi(team)
	if err != nil {
		h.errorLog.Println("ERROR_CreateMember_02: invalid form data:", err)
		utils.BadRequest(w, errors.New("invalid team id"))
		return
	}
	// 3. STEP ONE: Save data to Database (to generate ID)
	newMember := &models.Member{
		Name:           name,
		TeamID:         int64(teamID),
		Designation:    designation,
		Contact:        contact,
		Note:           note,
		ImageLink:      "", // Empty initially
		ShowOnHomepage: showOnHomepage,
	}

	id, err := h.DB.MemberRepo.Create(r.Context(), newMember)
	if err != nil {
		h.errorLog.Println("ERROR_CreateMember_03: db create:", err)
		utils.ServerError(w, errors.New("failed to save member info"))
		return
	}

	// 4. STEP TWO: Save Image to File System
	file, header, err := r.FormFile("profileImage")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			// No image uploaded, just return success with the ID
			h.respondSuccess(w, id, "Member created (no image uploaded)")
			return
		}
		h.errorLog.Println("ERROR_CreateMember_04: retrieving file:", err)
		utils.BadRequest(w, errors.New("invalid image file"))
		return
	}
	defer file.Close()

	// Create filename pattern: id_name.ext (e.g., 15_Alice_Smith.jpg)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	safeName := strings.ReplaceAll(name, " ", "_")
	filename := fmt.Sprintf("%d_%s%s", id, safeName, ext)

	// Create directory if not exists
	storagePath := filepath.Join("data", "images")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		h.errorLog.Println("ERROR_CreateMember_05: mkdir:", err)
		utils.ServerError(w, errors.New("server storage error"))
		return
	}

	fullPath := filepath.Join(storagePath, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		h.errorLog.Println("ERROR_CreateMember_06: create file:", err)
		utils.ServerError(w, errors.New("failed to create image file"))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.errorLog.Println("ERROR_CreateMember_07: save file:", err)
		utils.ServerError(w, errors.New("failed to save image content"))
		return
	}

	// 5. STEP THREE: Update Database with Image Link
	err = h.DB.MemberRepo.UpdateImageLink(r.Context(), id, filename)
	if err != nil {
		// Log error but do not fail request since data & file are saved
		h.errorLog.Println("ERROR_CreateMember_08: update link:", err)
	}

	h.respondSuccess(w, id, "Member created and image saved successfully")
}

// GetAllMembers retrieves a list of all members, optionally filtered by team ID, designation, and limited.
func (h *MemberHandler) GetAllMembers(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// --- A. Extract and parse maxLimit ---
	var maxLimit int64
	maxLimitStr := queryParams.Get("max_limit")
	if maxLimitStr != "" {
		val, err := strconv.ParseInt(maxLimitStr, 10, 64)
		if err != nil {
			h.errorLog.Println("ERROR_GetAllMembers_01: invalid max_limit format:", err)
			utils.BadRequest(w, errors.New("Invalid format for 'max_limit'. Must be an integer."))
			return
		}
		maxLimit = val
	}

	//homepage=true if req come from homepage
	homepage := queryParams.Get("show_on_home")
	showOnHomepage, _ := strconv.ParseBool(homepage)

	// --- B. Extract and parse designations list ---
	var designations []string
	designationStr := queryParams.Get("designations")
	if designationStr != "" {
		designations = strings.Split(designationStr, ",")

		// Trim whitespace from each designation for cleaner database lookups
		for i, d := range designations {
			designations[i] = strings.TrimSpace(d)
		}
	}

	// --- C. Extract and parse teamID ---
	var teamID int64
	teamIDStr := queryParams.Get("team_id")
	if teamIDStr != "" {
		val, err := strconv.ParseInt(teamIDStr, 10, 64)
		if err != nil {
			h.errorLog.Println("ERROR_GetAllMembers_02: invalid team_id format:", err)
			utils.BadRequest(w, errors.New("Invalid format for 'team_id'. Must be an integer."))
			return
		}
		teamID = val
	}

	// 2. Call the repository method with all three parameters
	// NEW SIGNATURE: GetAll(ctx context.Context, teamID int64, designations []string, maxLimit int64) ([]*models.Member, error)
	members, err := h.DB.MemberRepo.GetAll(r.Context(), teamID, maxLimit, showOnHomepage, designations)

	if err != nil {
		h.errorLog.Println("ERROR_GetAllMembers_03: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve members"))
		return
	}

	// 3. Prepare and send response
	var response struct {
		Error   bool             `json:"error"`
		Message string           `json:"message"`
		Members []*models.Member `json:"members"`
	}
	response.Error = false
	response.Message = "Members fetched successfully"
	response.Members = members
	utils.WriteJSON(w, http.StatusOK, response)
}

// GetLeadershipMessages retrieves messages from company leaders.
func (h *MemberHandler) GetLeadershipMessages(w http.ResponseWriter, r *http.Request) {
	members, err := h.DB.MemberRepo.GetAll(r.Context(), 0, 0, false, []string{"CHAIRMAN", "CEO & MANAGING DIRECTOR"})
	if err != nil {
		h.errorLog.Println("ERROR_GetLeadershipMessages_01: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve chairman info"))
		return
	}
	var response struct {
		Error    bool             `json:"error"`
		Message  string           `json:"message"`
		Leaders []*models.Member `json:"leaders"`
	}
	response.Error = false
	response.Message = "Leadership messages retrieved successfully"
	response.Leaders = members
	utils.WriteJSON(w, http.StatusOK, response)
}

// GetMember retrieves a single member by ID.
func (h *MemberHandler) GetMember(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid member ID"))
		return
	}

	member, err := h.DB.MemberRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_GetMember_01: db error:", err)
		utils.BadRequest(w, errors.New("member not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, member)
}

// UpdateMember handles updating member details and allows re-uploading the image.
func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid member ID"))
		return
	}

	// 1. Fetch existing member to preserve data/paths
	existing, err := h.DB.MemberRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateMember_01: fetch error:", err)
		utils.BadRequest(w, errors.New("member not found"))
		return
	}

	// 2. Parse Multipart Form (10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.errorLog.Println("ERROR_UpdateMember_02: parse form:", err)
		utils.BadRequest(w, errors.New("invalid form data"))
		return
	}

	// 3. Update Text Fields if provided
	name := strings.TrimSpace(r.FormValue("name"))
	if name != "" {
		existing.Name = name
	}

	teamIDStr := r.FormValue("team")
	teamID, err := strconv.Atoi(teamIDStr)
	if teamIDStr != "" && err == nil {
		existing.TeamID = int64(teamID)
	}

	designation := strings.TrimSpace(r.FormValue("designation"))
	if designation != "" {
		existing.Designation = designation
	}

	contact := strings.TrimSpace(r.FormValue("contact"))
	if contact != "" {
		existing.Contact = contact
	}

	note := strings.TrimSpace(r.FormValue("note"))
	if note != "" {
		existing.Note = note
	}
	if strings.TrimSpace(r.FormValue("showOnHome")) == "1" {
		existing.ShowOnHomepage = true
	} else {
		existing.ShowOnHomepage = false
	}
	// --- File Change Tracking Variables ---
	storagePath := filepath.Join("data", "images") // Base path: data/images
	oldImageLink := existing.ImageLink             // Store original image link
	var backupFilePath string                      // Tracks if a successful backup file was created

	// 4. Handle Optional Image Update
	file, header, err := r.FormFile("profileImage")
	if err == nil {
		// New image uploaded
		defer file.Close()

		// 4.1 Backup the Old Image if one exists
		if oldImageLink != "" {
			oldFullPath := filepath.Join(storagePath, oldImageLink)

			// Create backup filename: e.g., "10_Member.jpg" -> "10_Member_backup.jpg"
			oldExt := filepath.Ext(oldImageLink)
			oldBase := strings.TrimSuffix(oldImageLink, oldExt)
			backupFileName := fmt.Sprintf("%s_backup%s", oldBase, oldExt)
			backupFilePath = filepath.Join(storagePath, backupFileName)

			// STEP 1: Perform Backup (Rename old file to backup file)
			if err := os.Rename(oldFullPath, backupFilePath); err != nil {
				// We proceed if the file doesn't exist, but log an error if rename fails for another reason.
				if !os.IsNotExist(err) {
					h.errorLog.Println("WARNING_UpdateMember_03a: Failed to backup old member image:", err)
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
			h.errorLog.Println("ERROR_UpdateMember_04: mkdir:", err)

			// Restore backup if mkdir failed
			if backupFilePath != "" {
				restorePath := filepath.Join(storagePath, oldImageLink)
				if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
					h.errorLog.Println("CRITICAL_UpdateMember_04b: Failed to restore backup image:", restoreErr)
				}
			}

			utils.ServerError(w, errors.New("storage error"))
			return
		}

		fullPath := filepath.Join(storagePath, filename)
		dst, err := os.Create(fullPath)
		if err != nil {
			h.errorLog.Println("ERROR_UpdateMember_05: create file:", err)

			// Restore backup if create file failed
			if backupFilePath != "" {
				restorePath := filepath.Join(storagePath, oldImageLink)
				if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
					h.errorLog.Println("CRITICAL_UpdateMember_05b: Failed to restore backup image:", restoreErr)
				}
			}

			utils.ServerError(w, errors.New("failed to save new image"))
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			h.errorLog.Println("ERROR_UpdateMember_06: copy file:", err)

			// STEP 3 (Part 1): Restore backup if saving new file failed
			if backupFilePath != "" {
				restorePath := filepath.Join(storagePath, oldImageLink)
				// Delete the partially written new file before restore attempt
				os.Remove(fullPath)
				if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
					h.errorLog.Println("CRITICAL_UpdateMember_06a: Failed to restore backup image:", restoreErr)
				}
			}

			utils.ServerError(w, errors.New("failed to write new image"))
			return
		}

		// Update model with the new, successfully saved image link
		existing.ImageLink = filename
	} else if err != http.ErrMissingFile {
		// Handle other errors besides just missing the file
		h.errorLog.Println("ERROR_UpdateMember_06b: form file:", err)
		utils.BadRequest(w, errors.New("failed to process image upload"))
		return
	}

	// 5. Update Database
	err = h.DB.MemberRepo.Update(r.Context(), existing)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateMember_07: db update:", err)

		// STEP 3 (Part 2): Restore original file if DB update failed, AND a backup was made
		if backupFilePath != "" {
			// 1. Delete the newly uploaded file (whose name is now in existing.ImageLink)
			newFullPath := filepath.Join(storagePath, existing.ImageLink)
			if err := os.Remove(newFullPath); err != nil && !os.IsNotExist(err) {
				h.errorLog.Printf("WARNING_UpdateMember_07b: Failed to clean up new image (%s): %v", newFullPath, err)
			}

			// 2. Restore backup file
			restorePath := filepath.Join(storagePath, oldImageLink)
			if restoreErr := os.Rename(backupFilePath, restorePath); restoreErr != nil {
				h.errorLog.Println("CRITICAL_UpdateMember_07c: Failed to restore backup image after DB failure:", restoreErr)
			}
		}

		utils.ServerError(w, errors.New("failed to update member"))
		return
	}

	// 6. DB Update Successful: Cleanup
	// STEP 4: Remove backup file
	if backupFilePath != "" {
		if err := os.Remove(backupFilePath); err != nil {
			h.errorLog.Println("WARNING_UpdateMember_08: Failed to remove backup image:", err)
		}
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool           `json:"error"`
		Message string         `json:"message"`
		Data    *models.Member `json:"data"`
	}{
		Error:   false,
		Message: "Member updated successfully",
		Data:    existing,
	})
}

// DeleteMember removes the member from the database.
func (h *MemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid member ID"))
		return
	}

	//fetch the member info
	member, err := h.DB.MemberRepo.GetByID(r.Context(), id)
	if err != nil {
		utils.BadRequest(w, errors.New("Member not found"))
		return
	}

	err = h.DB.MemberRepo.Delete(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_DeleteMember_01: db error:", err)
		utils.ServerError(w, errors.New("failed to delete member"))
		return
	}

	//silently delete the image from the filesystem
	os.Remove(filepath.Join("data", "images", member.ImageLink))

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "Member deleted successfully",
	})
}

func (h *MemberHandler) respondSuccess(w http.ResponseWriter, id int64, msg string) {
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
