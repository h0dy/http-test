package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
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
