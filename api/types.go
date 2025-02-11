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

type CreateUserRequest struct {
	Email string `json:"email"`
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
