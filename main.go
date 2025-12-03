package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/h0dy/http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db        *database.Queries
	platform  string
	jwtSecret string
	polkaKey  string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error in loading env file: %v", err)
	}

	// get env variables for api configuration
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("make sure you set up DB_URL")
	}
	platform := os.Getenv("PLATFORM") // in what mode the app is running (e.g. dev or prod)
	if platform == "" {
		log.Fatal("make sure you set up PLATFORM")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("make sure you set up JWT_SECRET")
	}
	polkaKey := os.Getenv("POLKA_KEY") // polka key is an api key for webhook
	if polkaKey == "" {
		log.Fatal("make sure you set up polkaKey")
	}

	db, err := sql.Open("postgres", dbURL) // open connection to database
	if err != nil {
		log.Fatalf("error in connecting to database %v", err)
	}
	defer db.Close()

	dbQueries := database.New(db) // register the generated functions for our database queries from sqlc
	apiCfg := &apiConfig{
		db:        dbQueries,
		platform:  platform,
		jwtSecret: jwtSecret,
		polkaKey:  polkaKey,
	}

	const port = "8080"
	const filepath = "."
	mux := http.NewServeMux() // NewServeMux method returns ServeMux
	// "ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL." from go documents
	serveApp := http.StripPrefix("/app/", http.FileServer(http.Dir(filepath)))

	mux.Handle("/app/", serveApp)                            // serves the homepage (HTML page)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset) // resets the database
	mux.HandleFunc("GET /api/healthz", handlerReadiness)     // checks if the server is running

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken) // refresh access token (JWT)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeToken)   // revoke refresh token

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetSingleChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerChirpyUpgrade) // webhook endpoint

	server := &http.Server{Addr: ":" + port, Handler: mux}

	log.Printf("serving on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}
