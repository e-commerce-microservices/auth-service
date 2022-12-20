package service

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/e-commerce-microservices/auth-service/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthService represent all authentication and authorization logic
type AuthService struct {
	jwt        jwtManager
	userClient pb.UserServiceClient
	pb.UnimplementedAuthServiceServer
}

// NewAuthService creates a new AuthService
//
// params:
// + userClient: UserService grpc client  instance
func NewAuthService(userClient pb.UserServiceClient) *AuthService {
	service := &AuthService{
		jwt:        newJwtManager(time.Hour),
		userClient: userClient,
	}
	return service
}

// Login request receive email and password then return access token and refresh token
func (auth *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Println("--> GET: login")
	user, err := auth.userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// create access token
	accessToken, err := auth.jwt.createAccessToken(newUserClaims(strconv.FormatInt(user.Id, 10), user.Role))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// create refresh token
	// store session

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: "",
		Message:      "login successfully",
	}, nil
}

// Register request receive email, user_name and password then create new user in db
func (auth *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.GeneralResponse, error) {
	log.Println("--> GET: register")
	createUserResp, err := auth.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:    req.Email,
		UserName: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	return &pb.GeneralResponse{
		Message: createUserResp.GetMessage(),
	}, nil
}

// GetUserClaims return UserClaims for authenticated user
func (auth *AuthService) GetUserClaims(ctx context.Context, _ *empty.Empty) (*pb.UserClaimsResponse, error) {
	log.Println("--> GET: user claims")
	// parse header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "can't metadata from header")
	}

	// get authorization header
	values := md["authorization"]
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
		UserRole: *claims.role.Enum(),
	}, nil
}
