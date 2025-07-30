package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)


func TestJWT(t *testing.T) {
	tokenSecretTest := "chirpy-test"

	userID1 := uuid.New()
	token1, _ := MakeJWT(userID1, tokenSecretTest, time.Hour)
	
	userID2 := uuid.New()
	token2, _ := MakeJWT(userID2, tokenSecretTest, time.Hour)
	

	cases := []struct{
		name string
		token string
		tokenSecret string
		userID uuid.UUID
		expectErr bool
	}{
		{
			name: "Valid token 1",
			token: token1,
			tokenSecret: tokenSecretTest,
			userID: userID1,
			expectErr: false,
		},
		{
			name: "Invalid token 1",
			token: token1,
			tokenSecret: "invalid-token",
			userID: userID1,
			expectErr: true,
		},
		{
			name: "Valid token 2",
			token: token2,
			tokenSecret: tokenSecretTest,
			userID: userID2,
			expectErr: false,
		},
		{
			name: "Invalid token secret 2",
			token: token2,
			tokenSecret: "notATokenSecret",
			userID: uuid.Nil,
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			userID, err := ValidateJWT(c.token, c.tokenSecret)
	
			if err != nil {
				if !c.expectErr {
					t.Errorf("unexpected error: %v\ncase: %v", err, c.name)
				}
				return 
			}
	
			if c.expectErr {
				t.Errorf("expected error but got none\ncase: %v", c.name)
				return
			}
			if userID != c.userID {
				t.Errorf("userID mismatch: got %v, expected %v\ncase: %v", userID, c.userID, c.name)
			}
		})
	}
	
}
