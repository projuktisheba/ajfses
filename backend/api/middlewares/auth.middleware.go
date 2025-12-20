package middlewares

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	// Assuming you use this library
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"github.com/projuktisheba/ajfses/backend/internal/utils"
)

// Define a custom type for context key to avoid collisions
type ContextKey string

// AuthJWT creates a middleware function to validate a JWT and inject claims into the context.
func AuthJWT(jwtConfig models.JWTConfig, errorLog *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Get Authorization Header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorLog.Println("ERROR_01_AuthJWT: Authorization header missing.")
				utils.Unauthorized(w, errors.New("authorization token required"))
				return
			}

			// 2. Check for "Bearer " Prefix
			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				errorLog.Println("ERROR_02_AuthJWT: Invalid Authorization header format.")
				utils.Unauthorized(w, errors.New("invalid authorization header format. Expected 'Bearer <token>'"))
				return
			}

			tokenString := headerParts[1]

			// 3. Verify and Extract Claims using VerifyJWT (which uses MapClaims)
			claims, err := utils.VerifyJWT(tokenString, jwtConfig)
			if err != nil {
				errorLog.Printf("ERROR_03_AuthJWT: JWT validation failed: %v", err)
				utils.Unauthorized(w, errors.New("invalid or expired token"))
				return
			}

			// 4. Role Check (Only Admin is allowed for the Admin endpoints)
			if claims.Role != "Admin" {
				errorLog.Printf("ERROR_04_AuthJWT: Access denied for role: %s", claims.Role)
				utils.Unauthorized(w, errors.New("access denied. Insufficient privileges"))
				return
			}

			// 5. Inject Claims into Context
			// Use the context key to store the authenticated user claims
			// The key must match the one used in the handler (ContextKey("authClaims"))
			ctx := context.WithValue(r.Context(), models.AuthClaimsContextKey, *claims)

			// 6. Serve the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
