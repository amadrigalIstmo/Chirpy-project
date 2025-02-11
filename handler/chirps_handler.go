package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
	"github.com/google/uuid"
)

const maxChirpLength = 140

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func (h *Handler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	var req api.ChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Body == "" {
		api.RespondWithError(w, http.StatusBadRequest, "Chirp body cannot be empty")
		return
	}

	if len(req.Body) > maxChirpLength {
		api.RespondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	if req.UserID == uuid.Nil {
		api.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	cleanedText := filterProfanity(req.Body)

	params := database.CreateChirpParams{
		Body:   cleanedText,
		UserID: req.UserID,
	}

	newChirp, err := h.db.CreateChirp(r.Context(), params)
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		api.RespondWithError(w, http.StatusInternalServerError, "Could not create chirp")
		return
	}

	api.RespondWithJSON(w, http.StatusCreated, newChirp)
}

func (h *Handler) GetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := h.db.GetChirps(r.Context())
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps")
		return
	}

	var response []api.ChirpResponse
	for _, chirp := range chirps {
		response = append(response, api.ChirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: chirp.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	api.RespondWithJSON(w, http.StatusOK, response)
}

func filterProfanity(text string) string {
	words := strings.Fields(text)
	for i, word := range words {
		cleanWord := removePunctuation(word)
		for _, badWord := range profaneWords {
			if strings.EqualFold(cleanWord, badWord) {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

func removePunctuation(word string) string {
	re := regexp.MustCompile(`[^\w]`)
	return re.ReplaceAllString(word, "")
}
