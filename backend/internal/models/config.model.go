package models

import "time"

// Response is the type for response
type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// JWT holds token data
type JWT struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	ExpiresAt int64     `json:"exp"`
	IssuedAt  int64     `json:"iat"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JWTConfig struct {
	SecretKey string
	Issuer    string
	Audience  string
	Algorithm string
	Expiry    time.Duration
	Refresh   time.Duration
}

type DBConfig struct {
	DSN          string
	DEVDSN       string
}

type Config struct {
	Host  string
	Port  int64
	Env   string
	Owner string
	JWT   JWTConfig
	DB    DBConfig
}
