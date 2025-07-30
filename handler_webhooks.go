package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/h0dy/http-server/internal/auth"
)

func (cfg *apiConfig) handlerChirpyUpgrade(w http.ResponseWriter, r *http.Request) {
	type reqBody struct{
		Event string `json:"event"`
		Data struct{
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	
	polkaApiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	if polkaApiKey != cfg.polkaKey {
		respondWithErr(w, http.StatusUnauthorized, "API Key is invalid", nil)
		return
	}

	data := reqBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "something went wrong", err)
		return
	}
	if data.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := cfg.db.UpgradeToChirpyRed(r.Context(), data.Data.UserId);  err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithErr(w, http.StatusNotFound, "Couldn't find the user", err)
			return
		}
		respondWithErr(w, http.StatusInternalServerError, "Couldn't update user", err)
	}
	w.WriteHeader(http.StatusNoContent)
}