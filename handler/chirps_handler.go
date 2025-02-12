package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/amadrigalIstmo/Chirpy-project/api"
	"github.com/amadrigalIstmo/Chirpy-project/internal/auth"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"
	"github.com/google/uuid"
)

// / CreateChirp maneja la creación de un chirp, validando la autenticación con JWT.
func (h *Handler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	// Obtener y validar el token JWT del header Authorization
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, h.jwtSecret)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// Decodificar el cuerpo de la solicitud
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validar y limpiar el chirp
	cleaned, err := validateChirp(params.Body)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Crear el chirp en la base de datos
	chirp, err := h.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
	})
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	// Responder con el chirp creado
	api.RespondWithJSON(w, http.StatusCreated, api.Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
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
