package authentication

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, expiresIn time.Duration) (string, error) {
	secret_string := os.Getenv("secret_string")
	if len(secret_string) == 0 {
		return "", errors.New("\"secret_string\" variable not exported.")
	}
	token := jwt.NewWithClaims( // Generate a token that expires after expires period.
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "blade_db",
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	)
	tokenIdentifier, err := token.SignedString([]byte(secret_string))
	if err != nil {
		return "", err
	}
	return tokenIdentifier, nil
}

// Returns the user id embedded in the
func ValidateJWT(tokenString string) (uuid.UUID, error) {
	claimer := jwt.RegisteredClaims{}
	secret_string := os.Getenv("secret_string")
	if len(secret_string) == 0 {
		return uuid.Nil, errors.New("\"secret_string\" variable not exported.")
	}
	token, err := jwt.ParseWithClaims(tokenString, &claimer, func(token *jwt.Token) (any, error) {
		return []byte(secret_string), nil
	})

	// Check if parsing failed BEFORE accessing token.Claims
	if err != nil {
		return uuid.UUID{}, err
	}

	// Now it's safe to access token.Claims
	idStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	key := headers.Get("Authorization")
	btPair := strings.Fields(key)
	if len(btPair) != 2 {
		return "", fmt.Errorf("(\"Bearer\", token) pair not found in http header.")
	}
	if btPair[0] != "Bearer" {
		return "", fmt.Errorf("Keyword \"Bearer\" not found.")
	}
	return btPair[1], nil
}
