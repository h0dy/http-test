package main

import (
	"net/http"
	"time"

	"github.com/h0dy/http-server/internal/auth"
)

func (cfg *apiConfig)handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	w.Header().Set("Content-Type", "application/json")
	
	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), headerToken)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err  != nil {
		respondWithErr(w, http.StatusBadRequest, "Couldn't generate token", err)
	}
	
	respondWithJson(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	if err := cfg.db.SetRevokedAtToken(r.Context(), headerToken); err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Couldn't revoke token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
