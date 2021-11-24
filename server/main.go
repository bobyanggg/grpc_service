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
	"gitlab.com/leopardx602/grpc_service/worker"
	"google.golang.org/grpc"
)

type Server struct {
	consumer *worker.Consumer
}

type ProductGRPC struct {
	Products      chan pb.UserResponse
	FinishRequest chan int
}

func (s *Server) GetUserInfo(in *pb.UserRequest, stream pb.UserService_GetUserInfoServer) error {
	log.Println("Search for", in.KeyWord)

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	// Search in the database.
	products, err := sql.Select(in.KeyWord)
	if err != nil {
		return err
	}
	//fmt.Println(products)
	var p ProductGRPC
	p.Products = make(chan pb.UserResponse, 200)
	p.FinishRequest = make(chan int, 1)
	newProducts := make(chan *sql.Product, 200)

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

		go worker.Queue(ctx, in.KeyWord, newProducts)

		go func() {
			for product := range newProducts {
				// Push the data to grpc output.
				p.Products <- pb.UserResponse{
					Name:       product.Name,
					Price:      int32(product.Price),
					ImageURL:   product.ImageURL,
					ProductURL: product.ProductURL,
				}
			}
			fmt.Println("-------CLOSE-----------")
			p.FinishRequest <- 1
			time.Sleep(time.Second)
		}()
	}

	//output (work for from database)
	for {
		select {
		case product := <-p.Products:
			err := stream.Send(&product)
			if err != nil {
				log.Fatal(err)
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

	server := &Server{}
	pb.RegisterUserServiceServer(grpcServer, server)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcConfig["port"]))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(listen)
}
