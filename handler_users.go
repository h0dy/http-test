package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	
	user, err := cfg.db.CreateUser(context.Background(), data.Email)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't create user in DB", err)
		return 
	}
	respondWithJson(w, http.StatusCreated, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})
}