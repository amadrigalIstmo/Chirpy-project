package handler

import (
	"encoding/json"
	"net/http"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/auth"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req api.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		api.RespondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	// Hash de la contraseña antes de almacenarla
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	// Crear estructura con los parámetros esperados
	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	// Crear usuario en la base de datos
	newUser, err := h.db.CreateUser(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}

	// Respuesta sin incluir la contraseña
	response := api.CreateUserResponse{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	api.RespondWithJSON(w, http.StatusCreated, response)
}
