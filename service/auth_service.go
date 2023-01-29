package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/e-commerce-microservices/auth-service/pb"
	"github.com/e-commerce-microservices/auth-service/repository"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authRepository interface {
	CreateSession(ctx context.Context, arg repository.CreateSessionParams) error
	GetSession(ctx context.Context, refreshToken string) (repository.Session, error)
}

// AuthService represent all authentication and authorization logic
type AuthService struct {
	jwt        jwtManager
	authStore  authRepository
	userClient pb.UserServiceClient
	pb.UnimplementedAuthServiceServer
}

// NewAuthService creates a new AuthService
//
// params:
// + userClient: UserService grpc client  instance
func NewAuthService(authStore authRepository, userClient pb.UserServiceClient) *AuthService {
	service := &AuthService{
		jwt:        newJwtManager(),
		authStore:  authStore,
		userClient: userClient,
	}
	return service
}

// Login request receive email and password then return access token and refresh token
func (auth *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := auth.userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, errors.New("Email hoặc password không đúng, hãy thử lại")
	}

	// create access token
	userClaims := newUserClaims(strconv.FormatInt(user.Id, 10), user.Role)
	accessToken, err := auth.jwt.createAccessToken(userClaims)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// create refresh token
	refreshToken, err := auth.jwt.createRefreshToken(userClaims)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	mD, _ := metadata.FromIncomingContext(ctx)
	userAgent := ""
	if len(mD["user-agent"]) > 0 {
		userAgent = mD["user-agent"][0]
	}

	// store session
	err = auth.authStore.CreateSession(ctx, repository.CreateSessionParams{
		ID:           uuid.New(),
		UserID:       user.Id,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIp:     "",
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
	})
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Đăng nhập thành công",
	}, nil
}

// Register request receive email, user_name and password then create new user in db
func (auth *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.GeneralResponse, error) {
	createUserResp, err := auth.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    req.Email,
		UserName: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("Email %s đã được sử dụng, hãy thử đăng kí bằng email khác", req.GetEmail())
	}

	return &pb.GeneralResponse{
		Message: createUserResp.GetMessage(),
	}, nil
}

// GetUserClaims return UserClaims for authenticated user
func (auth *AuthService) GetUserClaims(ctx context.Context, _ *empty.Empty) (*pb.UserClaimsResponse, error) {
	// parse header
	// metadata.FromIncomingContext returns the incoming metadata in ctx if it exists
	// All keys in the returned MD are lowercase
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "can't read metadata from header")
	}

	values := md[authorizationHeader]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "can't parse authorization header")
	}

	accessToken := values[0]
	claims, err := auth.jwt.parseToken(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return &pb.UserClaimsResponse{
		Id:       claims.Id,
		UserRole: claims.Role,
	}, nil
}

// CustomerAuthorization return UserClaims for request from Customer
func (auth *AuthService) CustomerAuthorization(ctx context.Context, _ *empty.Empty) (*pb.UserClaimsResponse, error) {
	claims, err := auth.GetUserClaims(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	if claims.UserRole < pb.UserRole_customer {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return claims, nil
}

// SupplierAuthorization return UserClaims for request from Supplier
func (auth *AuthService) SupplierAuthorization(ctx context.Context, _ *empty.Empty) (*pb.UserClaimsResponse, error) {
	claims, err := auth.GetUserClaims(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	if claims.UserRole < pb.UserRole_supplier {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return claims, nil
}

// AdminAuthorization return UserClaims for request from Admin
func (auth *AuthService) AdminAuthorization(ctx context.Context, _ *empty.Empty) (*pb.UserClaimsResponse, error) {
	claims, err := auth.GetUserClaims(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	if claims.UserRole != pb.UserRole_admin {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return claims, nil
}

// Refresh ...
func (auth *AuthService) Refresh(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	claims, err := auth.jwt.parseToken(req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	// check session
	session, err := auth.authStore.GetSession(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid session")
	}
	if strconv.FormatInt(session.UserID, 10) != claims.Id {
		return nil, errors.New("invalid session")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("expired session")
	}

	user, err := auth.userClient.GetUserById(ctx, &pb.GetUserByIDRequest{
		UserId: session.UserID,
	})
	if err != nil {
		return nil, errors.New("invalid session")
	}
	// create user access token
	userClaims := newUserClaims(strconv.FormatInt(user.Id, 10), user.Role)
	accessToken, err := auth.jwt.createAccessToken(userClaims)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
		Message:      "Tạo thành công accessToken",
	}, nil
}

// Ping pong
func (auth *AuthService) Ping(context.Context, *empty.Empty) (*pb.Pong, error) {
	return &pb.Pong{
		Message: "pong",
	}, nil
}
