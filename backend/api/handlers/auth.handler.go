package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"github.com/projuktisheba/ajfses/backend/internal/utils"
)

type AuthHandler struct {
	DB        *dbrepo.DBRepository
	JWTConfig models.JWTConfig
	infoLog   *log.Logger
	errorLog  *log.Logger
}

func newAuthHandler(db *dbrepo.DBRepository, JWTConfig models.JWTConfig, infoLog, errorLog *log.Logger) AuthHandler {
	return AuthHandler{
		DB:        db,
		JWTConfig: JWTConfig,
		infoLog:   infoLog,
		errorLog:  errorLog,
	}
}

func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	type signinRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req signinRequest
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_Signin: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// Trim spaces
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		utils.BadRequest(w, errors.New("username and password are required"))
		return
	}

	// Fetch employee by mobile OR email
	user, err := h.DB.UserRepo.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		h.errorLog.Println("ERROR_02_Signin: user not found:", err)
		utils.BadRequest(w, errors.New("invalid username or password"))
		return
	}

	// Check password (assumes hashed in DB)
	if !utils.CheckPassword(req.Password, user.Password) {
		h.errorLog.Println("ERROR_03_Signin: password mismatch")
		utils.BadRequest(w, errors.New("invalid username or password"))
		return
	}

	// Check role (Only admin has the access)
	if user.Role != "Admin" {
		h.errorLog.Println("ERROR_03_Signin: Access Denied")
		utils.BadRequest(w, errors.New("You don't have the access. Please contact to the system administrator"))
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(models.JWT{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, h.JWTConfig)

	if err != nil {
		h.errorLog.Println("ERROR_04_Signin: failed to generate JWT:", err)
		utils.BadRequest(w, fmt.Errorf("failed to generate token: %w", err))
		return
	}

	// Build response
	resp := struct {
		Error        bool         `json:"error"`
		AccessToken  string       `json:"accessToken"`
		RefreshToken string       `json:"refreshToken"`
		User         *models.User `json:"user"`
	}{
		Error:       false,
		AccessToken: token,
		User:        user,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// UpdatePassword handles the updating of the authenticated user's password.
// It securely retrieves the user ID from the JWT claims in the request context.
func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {

	// 1. Define Request Body Structure and read JSON
	// (Existing code for reading NewPassword is fine)
	type passwordRequest struct {
		NewPassword string `json:"new_password"`
	}

	var req passwordRequest
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_UpdatePassword: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// Trim space and validate
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if len(req.NewPassword) < 6 {
		utils.BadRequest(w, errors.New("new password must be at least 6 characters long"))
		return
	}

	// 2. Get User ID SECURELY from Context
	// We expect the JWT middleware to have injected the claims using the key "authClaims".
	authClaims, ok := r.Context().Value(models.AuthClaimsContextKey).(models.JWT)
	if !ok {
		h.errorLog.Println("ERROR_02_UpdatePassword: authentication claims not found in context.")
		// If claims are missing, the user is not authenticated or middleware failed.
		utils.Unauthorized(w, errors.New("authentication context missing. Please log in again."))
		return
	}

	// *** SECURELY USE THE ID FROM THE CLAIMS ***
	userID := authClaims.ID
	// If you need to enforce that an Admin can reset OTHER user's passwords,
	// you would use the Claims Role + an ID from the URL/Body, but for a self-reset, this is correct.

	// 3. Hash the New Password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		h.errorLog.Println("ERROR_03_UpdatePassword: failed to hash password:", err)
		utils.ServerError(w, errors.New("internal server error during password hashing"))
		return
	}

	// 4. Update the Password in the Database
	err = h.DB.UserRepo.UpdatePassword(r.Context(), userID, hashedPassword)
	if err != nil {
		h.errorLog.Println("ERROR_04_UpdatePassword: database update failed:", err)
		utils.ServerError(w, errors.New("failed to update password in database"))
		return
	}

	// 5. Success Response
	resp := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "Password updated successfully.",
	}

	h.infoLog.Printf("User ID %d successfully updated their password.", userID)
	utils.WriteJSON(w, http.StatusOK, resp)
}
