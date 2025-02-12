package handler

import (
	"encoding/json"
	"net/http"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/auth"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
)

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Definir estructura para los parámetros esperados
	var req api.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Obtener el token JWT desde los encabezados
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	// Validar el JWT y obtener el ID del usuario autenticado
	userID, err := auth.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// Hashear la nueva contraseña
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Actualizar usuario en la base de datos
	updatedUser, err := h.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	// Responder con los datos actualizados del usuario (sin la contraseña)
	api.RespondWithJSON(w, http.StatusOK, api.UpdateUserResponse{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	})
}
