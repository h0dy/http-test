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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	
	data := body{}
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
	})
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds int `json:"expires_in_seconds"`
	}
	type response struct {
		User
		Token string `json:"token"`
	}
	
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	
	data := body{}
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
	
	if hour := 60 * 60; data.ExpiresInSeconds <= 0 || data.ExpiresInSeconds > hour {
		data.ExpiresInSeconds = hour
	}
	
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(data.ExpiresInSeconds) * time.Second)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "error in generating token", err)
		return
	}
	respondWithJson(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
		Token: token,
	})
}