package api

import (
	"time"

	"github.com/google/uuid"
)

// Chirp representa la estructura principal de un chirp
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

// Parameters representa los parámetros para crear un chirp
type ChirpCreationParams struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

// Request para crear usuario
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Request para actualizar usuario
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Request para login
type LoginRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"` // Campo opcional para el tiempo de expiración
}

// Respuesta sin incluir la contraseña
type CreateUserResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

// Respuesta del login que incluye el token
type LoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

// Response con los datos actualizados del usuario
type UpdateUserResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}
