package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/h0dy/http-server/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID  `json:"user_id"`
}

func validateChirp(body string) (string, error) {
	if body == "" {
		return "", errors.New("make sure to provide a body (chirp)")
	}
	if len(body) > 140 {
		return "", errors.New("chirp is too long! that's a premium feature")
	}
	return replaceProfaneWords(body), nil
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	w.Header().Set("Content-Type", "application/json")

	data := body{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't decode the json data", err)
		return
	}
	if data.UserId.String() == "" {
		respondWithErr(w, http.StatusBadRequest, "Make sure to provide a user id", nil)
	}
	cleaned_body, err := validateChirp(data.Body)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	
	chirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		Body: cleaned_body,
		UserID:data.UserId,
	})
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJson(w, http.StatusCreated, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}

func (cfg *apiConfig)handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirps, err  := cfg.db.GetAllChirps(context.Background())
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}
	chirpsJson := []Chirp{}
	for _, chirp := range chirps {
		chirpsJson = append(chirpsJson, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}
	respondWithJson(w, http.StatusOK, chirpsJson)

}

func (cfg *apiConfig)handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Invalid chirp id", err)
		return
	}
	chirp, err := cfg.db.GetChirp(context.Background(), chirpId)
	if err != nil {
		respondWithErr(w, http.StatusNotFound, "Couldn't retrieve the chirp", err)
		return
	}
	respondWithJson(w, http.StatusOK, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	})
}