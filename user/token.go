package user

import (
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/config"
	"github.com/golang-jwt/jwt/v5"
)

func (u *User) GetJWT() (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(7 * 24 * time.Hour)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  u.ID,
		"iat": now.Unix(),
		"exp": expiry.Unix(),
	}).SignedString(config.GetSigningSecret())
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiry, nil
}

func GetTokenID(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return config.GetSigningSecret(), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("Invalid token")
	}
	id, ok := claims["id"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid token")
	}

	return id, nil
}
