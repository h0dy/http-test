package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithErr(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	type errRes struct {
		Error string `json:"error"`
	}
	respondWithJson(w, code, errRes{
		Error: msg,
	})
}

func respondWithJson(w http.ResponseWriter, code int, payload any){
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling json: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}