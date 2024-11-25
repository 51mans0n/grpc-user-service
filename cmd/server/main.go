package main

import (
	"context"
	"fmt"
	"github.com/51mans0n/grpc-user-service/api/proto/userpb"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type userServer struct {
	userpb.UnimplementedUserServiceServer
	mu     sync.Mutex
	users  map[int64]*userpb.User
	nextID int64
}

func newServer() *userServer {
	return &userServer{
		users: make(map[int64]*userpb.User),
	}
}

func (s *userServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	user := &userpb.User{
		Id:    s.nextID,
		Name:  req.GetName(),
		Email: req.GetEmail(),
	}
	s.users[user.Id] = user

	log.Printf("Created user: %v", user)
	return &userpb.CreateUserResponse{User: user}, nil
}

func (s *userServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[req.GetId()]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", req.GetId())
	}

	log.Printf("Retrieved user: %v", user)
	return &userpb.GetUserResponse{User: user}, nil
}

func (s *userServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[req.GetId()]
	if !exists {
		return &userpb.DeleteUserResponse{Success: false}, fmt.Errorf("user with ID %d not found", req.GetId())
	}

	delete(s.users, req.GetId())
	log.Printf("Deleted user with ID: %d", req.GetId())
	return &userpb.DeleteUserResponse{Success: true}, nil
}

func main() {
	// Загрузка сертификатов
	certFile := "../../certs/server.crt"
	keyFile := "../../certs/server.key"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}

	// Создание gRPC-сервера с TLS
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	userpb.RegisterUserServiceServer(grpcServer, newServer())

	// Прослушивание на порту 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("Server is running on port 50051 with TLS")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
