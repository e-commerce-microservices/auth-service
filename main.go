package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/e-commerce-microservices/auth-service/pb"
	"github.com/e-commerce-microservices/auth-service/repository"
	"github.com/e-commerce-microservices/auth-service/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// postgres driver
	_ "github.com/lib/pq"
)

func main() {
	// create grpc server
	grpcServer := grpc.NewServer()

	// init user db connection
	pgDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWD"), os.Getenv("DB_DBNAME"),
	)

	authDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer authDB.Close()
	if err := authDB.Ping(); err != nil {
		log.Fatal("can't ping to user db", err)
	}

	// init user queries
	authQueries := repository.New(authDB)

	// dial user service
	userServiceConn, err := grpc.Dial("user-service:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("can't dial user service: ", err)
	}
	// create user service
	userService := pb.NewUserServiceClient(userServiceConn)
	// create auth service
	authService := service.NewAuthService(authQueries, userService)
	// regist auth service
	pb.RegisterAuthServiceServer(grpcServer, authService)

	// listen and server on tcp
	listener, err := net.Listen("tcp", ":8080")
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
