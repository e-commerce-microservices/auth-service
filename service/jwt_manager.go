package service

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type jwtManager struct {
	hmacSecret    string
	tokenDuration time.Duration
}

func newJwtManager(duration time.Duration) jwtManager {
	return jwtManager{
		hmacSecret:    os.Getenv("HMAC_SECRET"),
		tokenDuration: duration,
	}
}

func (jm jwtManager) createAccessToken(claims userClaims) (string, error) {
	claims.ExpiresAt = time.Now().Add(jm.tokenDuration).Unix()

	token := jwt.NewWithClaims(jm.secret.getMethod(), &claims)

	rawToken, err := token.SignedString(jm.secret.getSecretKey())
	// error when invalid header
	if err != nil {
		log.Printf("error occur when create new access token: %v", err)
	}
	return rawToken, err
}

func (jm JwtManager) parseToken(rawToken string) (*userClaims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jm.secret.getSecretKey(), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*userClaims)
	if !ok {
		return nil, errors.New("invalid claim")
	}

	return claims, nil
}

func (jm JwtManager) createRefreshToken() {
	panic("not implemented")
}
