package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type body struct {
		Body string `json:"body"`
	}
	type resJson struct {
		Cleaned string `json:"cleaned_body"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	data := body{}
	if err := decoder.Decode(&data); err != nil {
		respondWithErr(w, http.StatusInternalServerError, "Couldn't decode the body", err)
		return
	}

	if len(data.Body) >= 140 {
		respondWithErr(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleaned_body := replaceProfaneWords(data.Body)
	respondWithJson(w, http.StatusOK, resJson{Cleaned: cleaned_body})
}
