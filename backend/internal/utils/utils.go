package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/projuktisheba/ajfses/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// EnsureDir checks if a directory exists, and creates it if it does not.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}

// GenerateJWT generates a JWT token for the given user
func GenerateJWT(user models.JWT, cfg models.JWTConfig) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"id":         user.ID,
		"name":       user.Name,
		"username":   user.Username,
		"role":       user.Role,
		"iss":        cfg.Issuer,
		"aud":        cfg.Audience,
		"exp":        now.Add(cfg.Expiry).Unix(),
		"iat":        now.Unix(),
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(cfg.Algorithm), claims)
	return token.SignedString([]byte(cfg.SecretKey))
}

// ParseJWT validates the token and returns claims
func ParseJWT(tokenString string, cfg models.JWTConfig) (*models.JWT, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != cfg.Algorithm {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return &models.JWT{
		ID:        int64(claims["id"].(float64)),
		Name:      claims["name"].(string),
		Username:  claims["username"].(string),
		Role:      claims["role"].(string),
		Issuer:    claims["iss"].(string),
		Audience:  claims["aud"].(string),
		ExpiresAt: int64(claims["exp"].(float64)),
		IssuedAt:  int64(claims["iat"].(float64)),
	}, nil
}

// VerifyJWT validates the token string and returns the claims (models.JWT).
// It performs strict checks against the provided JWTConfig and safely parses time fields.
func VerifyJWT(tokenString string, cfg models.JWTConfig) (*models.JWT, error) {
	
	// 1. Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing algorithm matches the configuration
		if token.Method.Alg() != cfg.Algorithm {
			return nil, errors.New("unexpected signing method")
		}
		// Return the secret key for validation
		return []byte(cfg.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token is expired")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token signature or claims")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims format")
	}

	// 2. Perform Security/Config Checks and map claims
	
	// A. Check Issuer (iss)
	if claims["iss"].(string) != cfg.Issuer {
		return nil, errors.New("token issuer mismatch")
	}

	// B. Check Audience (aud)
	if claims["aud"].(string) != cfg.Audience {
		return nil, errors.New("token audience mismatch")
	}
    
	// C. Check Expiration (exp)
    // JSON numbers are parsed as float64
	expTimestamp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("token expiry claim missing or invalid")
	}
	if time.Unix(int64(expTimestamp), 0).Before(time.Now()) {
		return nil, errors.New("token has expired")
	}
	
	// 3. Map Claims to models.JWT with safe type conversions
	
	// Safely get ID (parsed as float64)
	id, ok := claims["id"].(float64) 
	if !ok {
		return nil, errors.New("token 'id' claim missing or invalid")
	}
    
    // Safely parse created_at time (FIX for the panic)
    createdAtStr, ok := claims["created_at"].(string)
    if !ok {
        return nil, errors.New("token 'created_at' claim missing or invalid format")
    }
    createdAt, err := time.Parse(time.RFC3339, createdAtStr)
    if err != nil {
        return nil, fmt.Errorf("failed to parse created_at time: %w", err)
    }

    // Safely parse updated_at time (FIX for the panic)
    updatedAtStr, ok := claims["updated_at"].(string)
    if !ok {
        return nil, errors.New("token 'updated_at' claim missing or invalid format")
    }
    updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
    if err != nil {
        return nil, fmt.Errorf("failed to parse updated_at time: %w", err)
    }


	return &models.JWT{
		ID:    int64(id),
		Name:   claims["name"].(string),
		Username: claims["username"].(string),
		Role:   claims["role"].(string),
		
		// Standard claims
		Issuer:  claims["iss"].(string),
		Audience: claims["aud"].(string),
		ExpiresAt: int64(claims["exp"].(float64)),
		IssuedAt: int64(claims["iat"].(float64)),
        
        // Time fields (Now correctly parsed from string)
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a plain password with its hashed version
func CheckPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

// Today returns the current date with time set to 00:00:00
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// NullableTime converts zero time to nil
func NullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func GetBranchID(r *http.Request) int64 {
	branchID, _ := strconv.ParseInt(r.Header.Get("X-Branch-ID"), 10, 64)
	return branchID
}

// GenerateMemoNo generates a memo number like "MMDD-4CHAR"
func GenerateMemoNo() string {
	// create a new rand with its own seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// MMDD part
	datePart := time.Now().Format("0102")
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 4 random alphanumeric characters
	randomPart := make([]byte, 4)
	for i := range randomPart {
		randomPart[i] = charset[r.Intn(len(charset))]
	}

	return fmt.Sprintf("%s%s", datePart, string(randomPart))
}
