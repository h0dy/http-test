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
	db *database.Queries
}

func main() {
	if err := godotenv.Load();  err != nil {
		log.Fatalf("error in loading env file: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("make sure to set up DB_URL")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error in connecting to database %v", err)
	}
	
	dbQueries := database.New(db)
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
	}

	const port = "8080"
	const filepath = "."
	mux := http.NewServeMux()
	serveApp := http.StripPrefix("/app/", http.FileServer(http.Dir(filepath)))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(serveApp))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("serving on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}
