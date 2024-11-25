package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/51mans0n/grpc-user-service/proto/userpb"
)

func main() {
	// Загрузка корневого сертификата
	certFile := "certs/server.crt"
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Fatalf("Failed to read certificate: %v", err)
	}

	// Создание пула сертификатов
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		log.Fatalf("Failed to add server certificate to certificate pool")
	}

	// Настройка TLS
	creds := credentials.NewClientTLSFromCert(certPool, "")

	// Установка соединения с TLS
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Создаём нового пользователя
	createResp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	})
	if err != nil {
		log.Fatalf("Could not create user: %v", err)
	}
	userID := createResp.GetUser().GetId()
	fmt.Printf("Created user: %v\n", createResp.GetUser())

	// Получаем информацию о пользователе
	getResp, err := client.GetUser(ctx, &pb.GetUserRequest{
		Id: userID,
	})
	if err != nil {
		log.Fatalf("Could not get user: %v", err)
	}
	fmt.Printf("Retrieved user: %v\n", getResp.GetUser())

	// Удаляем пользователя
	deleteResp, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{
		Id: userID,
	})
	if err != nil {
		log.Fatalf("Could not delete user: %v", err)
	}
	fmt.Printf("Deleted user success: %v\n", deleteResp.GetSuccess())
}
