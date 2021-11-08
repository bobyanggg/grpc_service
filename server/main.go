package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"gitlab.com/leopardx602/grpc_service/model"
	pb "gitlab.com/leopardx602/grpc_service/product"

	"google.golang.org/grpc"
)

const (
	port                  = 8081
	timeout time.Duration = 10
)

type Server struct {
}

func (s *Server) GetUserInfo(in *pb.UserRequest, stream pb.UserService_GetUserInfoServer) error {
	log.Println("Search for", in.KeyWord)

	// Check the database

	var p model.Product
	p.Products = make(chan pb.UserResponse, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Search
	go p.AddTest()

	// Send
	for {
		select {
		case product := <-p.Products:
			err := stream.Send(&product)
			if err != nil {
				log.Fatal(err)
				return err
			}
		case <-p.FinishRequest:
			log.Println("Done!")
			return nil
		case <-ctx.Done():
			log.Println("Time out")
			return nil
		}
	}
}

func main() {
	// GRPC service
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &Server{})
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(listen)
}
