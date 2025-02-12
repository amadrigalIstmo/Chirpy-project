package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/amadrigalIstmo/Chirpy-project/handler"
	"github.com/amadrigalIstmo/Chirpy-project/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB        *database.Queries
	Platform  string
	JWTSecret string // 🔹 Agregamos el campo para el secreto JWT
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in the environment variables")
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

	// 🔹 Pasamos jwtSecret al crear el Handler
	handlers := handler.NewHandler(apiCfg.DB, apiCfg.Platform, jwtSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/users", handlers.CreateUser)
	mux.HandleFunc("POST /api/chirps", handlers.CreateChirp)
	mux.HandleFunc("GET /api/chirps", handlers.GetChirps)
	mux.HandleFunc("POST /admin/reset", handlers.ResetDatabase)
	mux.HandleFunc("GET /api/chirps/{chirpID}", handlers.GetChirpByID)
	mux.HandleFunc("POST /api/login", handlers.Login)
	mux.HandleFunc("POST /api/refresh", handlers.RefreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", handlers.RevokeTokenHandler)

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
