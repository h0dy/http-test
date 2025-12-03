package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "chirpy-access"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{ // https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}
	// create a token with specified signing method and claims
	// SigningMethodHS256 is a signing method and its key is token of type []byte
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString([]byte(tokenSecret)) // returns complete JWT string
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// validate the signature of JWT and extract the claims into a (token *jwt.Token) struct
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims) // to get access to Claims
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token claims")
	}

	if claims.Issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userId, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid id; couldn't parse use id: %v", err.Error())
	}
	return userId, nil
}

// MarkRefreshToken func returns a random 256-bit string
func MarkRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}
