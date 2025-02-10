package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Estructura para manejar el estado del servidor
type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}

	// Endpoint de readiness "/healthz" (SOLO GET)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Servir archivos estáticos con métricas en "/app/"
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Endpoint "/metrics" para ver los hits (SOLO GET)
	mux.HandleFunc("GET /metrics", apiCfg.metricsHandler)

	// Endpoint "/reset" para reiniciar el contador (SOLO POST)
	mux.HandleFunc("POST /reset", apiCfg.resetHandler)

	// Servidor HTTP
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Servidor corriendo en http://localhost:8080")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}

// Middleware para contar hits en /app/
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Handler para mostrar las métricas en "/metrics"
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())
}

// Handler para resetear el contador en "/reset"
func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}
