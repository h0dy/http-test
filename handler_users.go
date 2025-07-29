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
	}

	data := body{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't decode the json data", err)
		return
	}
	user, err := cfg.db.GetUserByEmail(context.Background(), data.Email)
	if err != nil {
		respondWithErr(w, http.StatusBadRequest, "Something went wrong, or the user doesn't exists", err)
		return 
	}
	if err := auth.CheckPasswordHash(data.Password, user.HashedPassword); err != nil {
		respondWithErr(w, http.StatusUnauthorized, "Incorrect credential; incorrect email or password", err)
		return
	}

	respondWithJson(w, http.StatusOK, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}