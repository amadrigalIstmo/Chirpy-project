package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
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

// PolkaGetChirps maneja la obtención de chirps con filtro opcional por author_id.
// PolkaGetChirps maneja la obtención de chirps con filtro opcional por author_id y ordenamiento por created_at.
func (h *Handler) PolkaGetChirps(w http.ResponseWriter, r *http.Request) {
	// Obtener el parámetro opcional "author_id"
	authorIDString := r.URL.Query().Get("author_id")

	var authorID uuid.UUID
	var filterByAuthor bool

	if authorIDString != "" {
		parsedID, err := uuid.Parse(authorIDString)
		if err != nil {
			api.RespondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
			return
		}
		authorID = parsedID
		filterByAuthor = true
	}

	// Obtener todos los chirps de la base de datos
	dbChirps, err := h.db.GetChirps(r.Context())
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	// Obtener el parámetro opcional "sort"
	sortDirection := r.URL.Query().Get("sort")
	if sortDirection != "desc" { // "asc" es el valor por defecto
		sortDirection = "asc"
	}

	// Filtrar chirps si se proporcionó un author_id
	chirps := []api.Chirp{}
	for _, dbChirp := range dbChirps {
		if filterByAuthor && dbChirp.UserID != authorID {
			continue
		}

		chirps = append(chirps, api.Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	// Ordenar los chirps según el parámetro "sort"
	sort.Slice(chirps, func(i, j int) bool {
		if sortDirection == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	api.RespondWithJSON(w, http.StatusOK, chirps)
}

func (h *Handler) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	// Extraer el `chirpID` de la URL
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

	// Extraer el `userID` desde el token JWT en los headers
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

	// Obtener el chirp desde la base de datos para verificar el dueño
	chirp, err := h.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	// Verificar que el usuario autenticado sea el dueño del chirp
	if chirp.UserID != userID {
		api.RespondWithError(w, http.StatusForbidden, "You are not the owner of this chirp", nil)
		return
	}

	// Eliminar el chirp
	err = h.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	// Responder con éxito (204 No Content)
	w.WriteHeader(http.StatusNoContent)
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
