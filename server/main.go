package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"gitlab.com/leopardx602/grpc_service/model"
	pb "gitlab.com/leopardx602/grpc_service/product"
	"gitlab.com/leopardx602/grpc_service/sql"

	"google.golang.org/grpc"
)

type Server struct {
	//sql.conn()
}

type ProductGRPC struct {
	Products      chan pb.UserResponse
	FinishRequest chan int
}

func (s *Server) GetUserInfo(in *pb.UserRequest, stream pb.UserService_GetUserInfoServer) error {
	log.Println("Search for", in.KeyWord)

	// Search in the database.
	products, err := sql.Select(in.KeyWord)
	if err != nil {
		return err
	}

	var p ProductGRPC
	p.Products = make(chan pb.UserResponse, 200)
	p.FinishRequest = make(chan int, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if len(products) > 0 {
		// Push the data to grpc output.
		go func() {
			for _, product := range products {
				p.Products <- pb.UserResponse{
					Name:       product.Name,
					Price:      int32(product.Price),
					ImageURL:   product.ImageURL,
					ProductURL: product.ProductURL,
				}
			}
			for {
				if len(p.Products) == 0 {
					p.FinishRequest <- 1
					time.Sleep(time.Second)
				}
			}
		}()

	} else {
		// Search for keyword in webs
		newProducts := make(chan sql.Product, 200)

		//go findtest.FindInWeb(stream.Context(), newProducts, in.KeyWord)

		go func() {
			for product := range newProducts {
				fmt.Println(product)
				// Insert the data to the database.
				product.Word = in.KeyWord
				if err := sql.Insert(product); err != nil {
					log.Println(err)
				}

				// Push the data to grpc output.
				p.Products <- pb.UserResponse{
					Name:       product.Name,
					Price:      int32(product.Price),
					ImageURL:   product.ImageURL,
					ProductURL: product.ProductURL,
				}
			}
			for {
				if len(p.Products) == 0 {
					p.FinishRequest <- 1
					time.Sleep(time.Second)
				}
			}
		}()
	}

	// Output
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
	// Read the grpc config.
	grpcConfig, err := model.OpenJson("../config/grpc.json")
	if err != nil {
		log.Fatal(err)
	}

	// GRPC service
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &Server{})
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcConfig["port"]))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(listen)
}
