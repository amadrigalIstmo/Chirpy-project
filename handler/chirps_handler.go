package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
	"github.com/google/uuid"
)

func (h *Handler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := api.ChirpCreationParams{}
	err := decoder.Decode(&params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: params.UserID,
	})
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	response := api.Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	api.RespondWithJSON(w, http.StatusCreated, response)
}

func (h *Handler) GetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := h.db.GetChirps(r.Context())
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps", err)
		return
	}

	var response []api.Chirp
	for _, chirp := range chirps {
		response = append(response, api.Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	api.RespondWithJSON(w, http.StatusOK, response)
}

func (h *Handler) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	if chirpIDStr == "" {
		api.RespondWithError(w, http.StatusBadRequest, "Chirp ID is required", nil)
		return
	}

	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
		return
	}

	chirp, err := h.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	response := api.Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	api.RespondWithJSON(w, http.StatusOK, response)
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
