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
	platform string
	jwtSecret string
	polkaKey string
}

func main() {
	if err := godotenv.Load();  err != nil {
		log.Fatalf("error in loading env file: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("make sure you set up DB_URL")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("make sure you set up PLATFORM")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("make sure you set up JWT_SECRET")
	}

	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("make sure you set up polkaKey")
	}
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error in connecting to database %v", err)
	}
	
	dbQueries := database.New(db)
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
		platform: platform,
		jwtSecret: jwtSecret,
		polkaKey: polkaKey,
	}

	const port = "8080"
	const filepath = "."
	mux := http.NewServeMux()
	serveApp := http.StripPrefix("/app/", http.FileServer(http.Dir(filepath)))

	mux.Handle("/app/", serveApp)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeToken)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetSingleChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerChirpyUpgrade)

	server := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("serving on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}
