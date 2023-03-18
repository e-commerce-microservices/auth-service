package service

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type jwtManager struct {
	hmacSecret []byte
}

func newJwtManager() jwtManager {
	return jwtManager{
		hmacSecret: []byte(os.Getenv("HMAC_SECRET")),
	}
}

func (jm jwtManager) createAccessToken(claims userClaims) (string, error) {
	claims.ExpiresAt = time.Now().Add(time.Hour * 30).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	rawToken, err := token.SignedString(jm.hmacSecret)
	// error when invalid header
	if err != nil {
		log.Printf("error occur when create new access token: %v", err)
	}
	return rawToken, err
}

func (jm jwtManager) parseToken(rawToken string) (*userClaims, error) {
	token, err := jwt.ParseWithClaims(rawToken, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
		// check algo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid algo header")
		}

		return jm.hmacSecret, nil
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

func (jm jwtManager) createRefreshToken(claims userClaims) (string, error) {
	claims.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	rawToken, err := token.SignedString(jm.hmacSecret)
	// error when invalid header
	if err != nil {
		log.Printf("error occur when create new refresh token: %v", err)
	}
	return rawToken, err
}
