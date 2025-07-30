package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/h0dy/http-server/internal/auth"
	"github.com/h0dy/http-server/internal/database"

	"github.com/google/uuid"
)
type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	
	data := reqBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't decode the json data", err)
		return
	}
	if data.Email == "" {
		respondWithErr(w, http.StatusBadRequest, "Make sure to provide the email", nil)
		return
	}
	if data.Password == "" {
		respondWithErr(w, http.StatusBadRequest, "Make sure to provide the password", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't hash the password", err)
	}
	
	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email: data.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Something went wrong, maybe try login in instead", err)
		return 
	}
	respondWithJson(w, http.StatusCreated, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	
	data := reqBody{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't decode the json data", err)
		return
	}
	user, err := cfg.db.GetUserByEmail(context.Background(), data.Email)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Incorrect credential; incorrect email or password", err)
		return 
	}
	
	if err := auth.CheckPasswordHash(data.Password, user.HashedPassword); err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Incorrect credential; incorrect email or password", err)
		return
	}
	
	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "error in generating token", err)
		return
	}
	
	refreshToken := auth.MarkRefreshToken()
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpireAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	respondWithJson(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token: accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	
	type response struct {
		User
		Token string `json:"token"`
	}

	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	data := reqBody{}
	decoder := json.NewDecoder(r.Body) 
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "something went wrong", err)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	
	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	
	hashedPassword, err := auth.HashPassword(data.Password)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't hash the password", err)
	}

	user, err := cfg.db.UpdateUserPassEmail(r.Context(), database.UpdateUserPassEmailParams{
		Email:data.Email,
		HashedPassword: hashedPassword,
		ID: userID,
	}); 
	if err != nil {
		respondWithErr(w, http.StatusNotFound, "Couldn't find the user", err)
		return
	}

	respondWithJson(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:accessToken,
	})
}