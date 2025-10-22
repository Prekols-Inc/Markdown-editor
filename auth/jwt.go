package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var (
	JWT_SECRET        = []byte(os.Getenv("JWT_SECRET"))
	ACCESS_TOKEN_TTL  = time.Minute * 15
	REFRESH_TOKEN_TTL = time.Hour * 24
)

func generateToken(userID uuid.UUID, TTL time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(TTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_SECRET)
}

func generateTokens(userID uuid.UUID) (string, string, error) {
	accessToken, err := generateToken(userID, ACCESS_TOKEN_TTL)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(userID, REFRESH_TOKEN_TTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return JWT_SECRET, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
