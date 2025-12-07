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

type TeamHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newTeamHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) TeamHandler {
	return TeamHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

// CreateTeam handles the creation of a new team.
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req models.Team
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_CreateTeam_01: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		utils.BadRequest(w, errors.New("title is required"))
		return
	}

	id, err := h.DB.TeamRepo.Create(r.Context(), &req)
	if err != nil {
		h.errorLog.Println("ERROR_CreateTeam_02: db create:", err)
		utils.ServerError(w, errors.New("failed to create team"))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}{
		Error:   false,
		Message: "Team created successfully",
		ID:      id,
	})
}

// GetAllTeams retrieves all teams.
func (h *TeamHandler) GetAllTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.DB.TeamRepo.GetAll(r.Context())
	if err != nil {
		h.errorLog.Println("ERROR_GetAllTeams_01: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve teams"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, teams)
}

// GetTeam retrieves a single team by ID.
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid team ID"))
		return
	}

	team, err := h.DB.TeamRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_GetTeam_01: db error:", err)
		utils.BadRequest(w, errors.New("team not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, team)
}

func (h *TeamHandler) GetAllTeamsAndMembers(w http.ResponseWriter, r *http.Request) {
	// Call the new Repo method
	teamsData, err := h.DB.MemberRepo.GetTeamsWithMembers(r.Context())
	if err != nil {
		h.errorLog.Println("ERROR_GetTeams_01: db error:", err)
		utils.ServerError(w, errors.New("failed to retrieve team data"))
		return
	}

	var response struct {
		Error   bool               `json:"error"`
		Message string             `json:"message"`
		Data    []*models.TeamData `json:"data"` // This holds the nested list
	}

	response.Error = false
	response.Message = "Teams and members fetched successfully"
	response.Data = teamsData

	utils.WriteJSON(w, http.StatusOK, response)
}

// UpdateTeam updates an existing team.
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid team ID"))
		return
	}

	// 1. Fetch existing team
	existing, err := h.DB.TeamRepo.GetByID(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateTeam_01: fetch error:", err)
		utils.BadRequest(w, errors.New("team not found"))
		return
	}

	// 2. Parse payload
	var req models.Team
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_UpdateTeam_02: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// 3. Update fields
	req.Title = strings.TrimSpace(req.Title)
	if req.Title != "" {
		existing.Title = req.Title
	}

	// 4. Save updates
	err = h.DB.TeamRepo.Update(r.Context(), existing)
	if err != nil {
		h.errorLog.Println("ERROR_UpdateTeam_03: db update:", err)
		utils.ServerError(w, errors.New("failed to update team"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool         `json:"error"`
		Message string       `json:"message"`
		Data    *models.Team `json:"data"`
	}{
		Error:   false,
		Message: "Team updated successfully",
		Data:    existing,
	})
}

// DeleteTeam removes a team.
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid team ID"))
		return
	}

	err = h.DB.TeamRepo.Delete(r.Context(), id)
	if err != nil {
		h.errorLog.Println("ERROR_DeleteTeam_01: db error:", err)
		utils.ServerError(w, errors.New("failed to delete team"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "Team deleted successfully",
	})
}
