package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/amadrigalIstmo/Chirpy-project/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const maxChirpLength = 140

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type apiConfig struct {
	DB       *database.Queries
	Platform string
}

type chirpRequest struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type createUserRequest struct {
	Email string `json:"email"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in the environment variables")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Could not ping database:", err)
	}

	apiCfg := apiConfig{
		DB:       database.New(db),
		Platform: os.Getenv("PLATFORM"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	//mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps) // Nueva ruta GET
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetDatabase)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Servidor corriendo en http://localhost:8080")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}

func (apiCfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	var req chirpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validar que el body no esté vacío
	if req.Body == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp body cannot be empty")
		return
	}

	// Validar la longitud del chirp
	if len(req.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	// Validar que el user_id no esté vacío
	if req.UserID == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Limpiar el texto de palabras prohibidas
	cleanedText := filterProfanity(req.Body)

	// Crear el chirp en la base de datos
	params := database.CreateChirpParams{
		Body:   cleanedText,
		UserID: req.UserID,
	}

	newChirp, err := apiCfg.DB.CreateChirp(r.Context(), params)
	if err != nil {
		log.Printf("Error creating chirp: %v", err) // Agregar log para debugging
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, newChirp)
}

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validar que el email no esté vacío
	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	newUser, err := apiCfg.DB.CreateUser(r.Context(), req.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, newUser)
}

func (apiCfg *apiConfig) handlerResetDatabase(w http.ResponseWriter, r *http.Request) {
	if apiCfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Action not allowed in production")
		return
	}

	err := apiCfg.DB.Reset(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not reset database")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Database reset successful"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
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
