package lib

import (
	"fmt"
	"time"

	"github.com/muhammadwahyupratamaa/assessment-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role models.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(
	userID string,
	role models.Role,
	secret string,
	expiresHours int,
) (string, error) {
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  userID,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Duration(expiresHours) * time.Hour),
			),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return []byte(secret), nil
		},
	)

	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token claims")
	}

	return *claims, nil
}
