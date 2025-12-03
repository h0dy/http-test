package main

import (
	"slices"
	"strings"
)

func replaceProfaneWords(input string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(input, " ")
	for idx, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			words[idx] = "****"
		}
	}
	cleaned_input := strings.Join(words, " ")
	return cleaned_input
}
