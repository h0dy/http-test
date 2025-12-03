package auth

import (
	"errors"
	"net/http"
	"strings"
)

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	apiKey := strings.TrimSpace(strings.TrimPrefix(authHeader, "ApiKey"))
	if apiKey == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	return apiKey, nil
}

// GetBearerToken func gets Authorization (header) value
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	authToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
	if authToken == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	return authToken, nil
}
