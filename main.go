package main

import (
	"log"
	"net"

	"github.com/e-commerce-microservices/auth-service/pb"
	"github.com/e-commerce-microservices/auth-service/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// create grpc server
	grpcServer := grpc.NewServer()

	// dial user service
	userServiceConn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("can't dial user service: ", err)
	}
	// create user service
	userService := pb.NewUserServiceClient(userServiceConn)
	// create auth service
	authService := service.NewAuthService(userService)
	// regist auth service
	pb.RegisterAuthServiceServer(grpcServer, authService)

	// enable reflection
	reflection.Register(grpcServer)

	// listen and server on tcp
	listener, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		log.Fatal("can't create listener: ", err)
	}

	log.Printf("start gRPC server on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot create grpc server: ", err)
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}
