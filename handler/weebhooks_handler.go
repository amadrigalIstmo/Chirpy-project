package handler

import (
	"encoding/json"
	"net/http"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/auth"
	"github.com/google/uuid"
)

// PolkaWebhook maneja los webhooks de Polka
func (h *Handler) PolkaWebhook(w http.ResponseWriter, r *http.Request) {
	// 🔹 Validamos la API Key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != h.polkaKey {
		api.RespondWithError(w, http.StatusUnauthorized, "Invalid API Key", nil)
		return
	}

	type WebhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	var req WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	// Ignorar eventos que no sean "user.upgraded"
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Validar el formato del UUID
	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format", err)
		return
	}

	// Actualizar el usuario a Chirpy Red
	_, err = h.db.UpgradeToChirpyRed(r.Context(), userID)
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	// Responder con 204 No Content si la actualización fue exitosa
	w.WriteHeader(http.StatusNoContent)
}
