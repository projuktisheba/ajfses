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
		Name:        name,
		TeamID:      int64(teamID),
		Designation: designation,
		Contact:     contact,
		Note:        note,
		ImageLink:   "", // Empty initially
	}
	fmt.Println(newMember)
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

// GetAllMembers retrieves a list of all members.
func (h *MemberHandler) GetAllMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.DB.MemberRepo.GetAll(r.Context(), "")
	if err != nil {
		h.errorLog.Println("ERROR_GetAllMembers_01: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve members"))
		return
	}
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

// GetAllMembers retrieves a list of all members.
func (h *MemberHandler) GetChairmanInfo(w http.ResponseWriter, r *http.Request) {
	members, err := h.DB.MemberRepo.GetAll(r.Context(), "Chairman")
	if err != nil {
		h.errorLog.Println("ERROR_GetChairmanInfo_01: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve chairman info"))
		return
	}
	var response struct {
		Error    bool             `json:"error"`
		Message  string           `json:"message"`
		Chairman []*models.Member `json:"Chairman"`
	}
	response.Error = false
	response.Message = "Chairman info retrieved successfully"
	response.Chairman = members
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
	idStr := chi.URLParam(r, "id")
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
	if teamIDStr != "" {
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

	// 4. Handle Optional Image Update
	file, header, err := r.FormFile("profileImage")
	if err == nil {
		// New image uploaded
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}
		safeName := strings.ReplaceAll(existing.Name, " ", "_")
		filename := fmt.Sprintf("%d_%s%s", id, safeName, ext)

		storagePath := filepath.Join("data", "images")
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			h.errorLog.Println("ERROR_UpdateMember_03: mkdir:", err)
			utils.ServerError(w, errors.New("storage error"))
			return
		}

		fullPath := filepath.Join(storagePath, filename)
		dst, err := os.Create(fullPath)
		if err != nil {
			h.errorLog.Println("ERROR_UpdateMember_04: create file:", err)
			utils.ServerError(w, errors.New("failed to save new image"))
			return
		}

		if _, err := io.Copy(dst, file); err != nil {
			dst.Close()
			h.errorLog.Println("ERROR_UpdateMember_05: copy file:", err)
			utils.ServerError(w, errors.New("failed to write new image"))
			return
		}
		dst.Close()

		existing.ImageLink = fullPath
	}

	// 5. Update Database
	err = h.DB.MemberRepo.Update(r.Context(), existing)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateMember_06: db update:", err)
		utils.ServerError(w, errors.New("failed to update member"))
		return
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
