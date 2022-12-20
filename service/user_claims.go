package service

import (
	"context"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userClaims struct {
	jwt.StandardClaims
	Role pb.User_Role `json:"role"`
}
type userClaimsKey = struct{}

func newUserClaims(userID string, role pb.User_Role) userClaims {
	return userClaims{
		StandardClaims: jwt.StandardClaims{
			Id: userID,
		},
		Role: role,
	}
}

func injectUserClaimsToContext(ctx context.Context, claims *userClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey{}, claims)
}

func extractUserClaimsFromContext(ctx context.Context) (*userClaims, error) {
	claims, ok := ctx.Value(userClaimsKey{}).(*userClaims)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "permission denied")
	}

	return claims, nil
}
