package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/h0dy/http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	if err := godotenv.Load();  err != nil {
		log.Fatal("error in loading env file")
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("error in connecting to database")
	}
	dbQueries := database.New(db)

	const port = "8080"
	const filepath = "."
	mux := http.NewServeMux()
	serveApp := http.StripPrefix("/app/", http.FileServer(http.Dir(filepath)))

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(serveApp))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("serving on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}
