package service

import (
	"github.com/e-commerce-microservices/auth-service/pb"
	"github.com/golang-jwt/jwt"
)

type userClaims struct {
	jwt.StandardClaims
	role pb.UserRole
}

func newUserClaims(userID string, role pb.UserRole) userClaims {
	return userClaims{
		StandardClaims: jwt.StandardClaims{
			Id: userID,
		},
	}
}
