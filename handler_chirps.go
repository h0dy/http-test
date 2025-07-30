package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/h0dy/http-server/internal/auth"
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
	type reqBody struct {
		Body string `json:"body"`
	}

	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	userId, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	data := reqBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't decode the json data", err)
		return
	}

	cleaned_body, err := validateChirp(data.Body)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleaned_body,
		UserID:userId,
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

	authorId, err := uuid.Parse(r.URL.Query().Get("author_id"))
	validAuthorId := true
	if err != nil {
		validAuthorId = false
	}
	
	chirps, err := cfg.db.GetAllChirps(r.Context(), uuid.NullUUID{
		UUID: authorId,
		Valid: validAuthorId,
	})
	sortChirps := r.URL.Query().Get("sort")
	if sortChirps == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}
	
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
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
	w.Header().Set("Content-Type", "application/json")
	
	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)
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

func (cfg *apiConfig)handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Invalid chirp id", err)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	userId, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		respondWithErr(w, http.StatusNotFound, "Chirp not found", err)
		return
	}
	if chirp.UserID != userId {
		respondWithErr(w, http.StatusForbidden, "chirp owner doesn't match access token's id", err)
		return
	}
	
	cfg.db.DeleteChirpById(r.Context(), chirpId)
	w.WriteHeader(http.StatusNoContent)
}
