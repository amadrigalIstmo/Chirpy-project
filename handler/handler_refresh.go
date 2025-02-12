package handler

import (
	"net/http"
	"time"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/auth"
)

// Handler para refrescar el token de acceso
func (h *Handler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "No se encontró el token", err)
		return
	}

	user, err := h.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "No se pudo obtener el usuario para el refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		h.jwtSecret,
		time.Hour, // Token de acceso válido por 1 hora
	)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "No se pudo generar un nuevo token de acceso", err)
		return
	}

	api.RespondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

// Handler para revocar un refresh token
func (h *Handler) RevokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "No se encontró el token", err)
		return
	}

	_, err = h.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "No se pudo revocar la sesión", err)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content (sin cuerpo)
}
