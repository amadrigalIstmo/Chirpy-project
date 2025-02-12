package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/auth"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
)

// Login permite a un usuario autenticarse y recibir un JWT
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Buscar usuario por email
	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	// Comparar contraseñas
	if err := auth.CheckPasswordHash(req.Password, user.HashedPassword); err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	// Generar Access Token (JWT) - válido por 1 hora
	accessToken, err := auth.MakeJWT(user.ID, h.jwtSecret, time.Hour)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not generate access token", err)
		return
	}

	// Generar Refresh Token - válido por 60 días
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not generate refresh token", err)
		return
	}

	// Guardar Refresh Token en la base de datos
	_, err = h.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60), // 60 días
	})
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not save refresh token", err)
		return
	}

	// Responder con usuario, access token y refresh token
	response := api.LoginResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}

	api.RespondWithJSON(w, http.StatusOK, response)
}
