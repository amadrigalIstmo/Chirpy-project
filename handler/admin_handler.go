package handler

import (
	"net/http"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
)

type Handler struct {
	db       *database.Queries
	platform string
}

func NewHandler(db *database.Queries, platform string) *Handler {
	return &Handler{
		db:       db,
		platform: platform,
	}
}

func (h *Handler) ResetDatabase(w http.ResponseWriter, r *http.Request) {
	if h.platform != "dev" {
		api.RespondWithError(w, http.StatusForbidden, "Action not allowed in production")
		return
	}

	err := h.db.Reset(r.Context())
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not reset database")
		return
	}

	api.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Database reset successful"})
}
