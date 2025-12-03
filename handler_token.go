package main

import (
	"net/http"
	"time"

	"github.com/h0dy/http-server/internal/auth"
)

// handlerRefreshToken func creates a new access token(JWT) with the refresh token in the header
func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	w.Header().Set("Content-Type", "application/json")

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		return
	}

	// create a new access token(JWT)
	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Couldn't generate token", err)
	}

	respondWithJson(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, err.Error(), err)
		return
	}
	if err := cfg.db.SetRevokedAtToken(r.Context(), refreshToken); err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Couldn't revoke token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// with cookies
// func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {

// 	cookie, err := r.Cookie("refresh_token")
// 	if err != nil {
// 		respondWithErr(w, http.StatusUnauthorized, "Missing refresh token cookie", err)
// 		return
// 	}
// 	refreshToken := cookie.Value

// 	type response struct {
// 		Token string `json:"token"`
// 	}
// 	w.Header().Set("Content-Type", "application/json")

// 	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
// 	if err != nil {
// 		respondWithErr(w, http.StatusUnauthorized, "Invalid or expired refresh token", err)
// 		return
// 	}

// 	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
// 	if err != nil {
// 		respondWithErr(w, http.StatusInternalServerError, "Couldn't generate new access token", err)
// 		return
// 	}

// 	newRefresh := auth.MarkRefreshToken()
// 	if err := cfg.db.UpdateRefreshToken(r.Context(), user.ID, newRefresh); err == nil {
// 		http.SetCookie(w, &http.Cookie{
// 			Name:     "refresh_token",
// 			Value:    newRefresh,
// 			HttpOnly: true,
// 			Secure:   true,
// 			SameSite: http.SameSiteStrictMode,
// 			Path:     "/api/refresh",
// 			MaxAge:   int((24 * time.Hour * 60).Seconds()),
// 		})
// 	}
// 	respondWithJson(w, http.StatusOK, response{
// 		Token: accessToken,
// 	})
// }
