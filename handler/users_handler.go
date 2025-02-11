package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/amadrigalIstmo/Chirpy-project/api"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req api.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if req.Email == "" {
		api.RespondWithError(w, http.StatusBadRequest, "Email is required", nil)
		return
	}

	newUser, err := h.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error al crear usuario: %v", err)
		api.RespondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}

	response := api.CreateUserResponse{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	api.RespondWithJSON(w, http.StatusCreated, response)
}
