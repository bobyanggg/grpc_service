package main

import (
	"fmt"
	"log"
	"net"

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

	// Search in the database.
	products, err := sql.Select(in.KeyWord)
	if err != nil {
		return err
	}
	//fmt.Println(products)
	var p ProductGRPC
	p.Products = make(chan pb.UserResponse, 200)
	p.FinishRequest = make(chan int, 1)

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
				select {
				case <-stream.Context().Done():
					log.Println("..........ctx canceled...........", stream.Context().Err())
					return
				default:
					if len(p.Products) == 0 {
						p.FinishRequest <- 1
						return
					}
				}

			}
		}()

	} else {
		// Search for keyword in webs
		go func() {
			worker.Queue(stream.Context(), in.KeyWord, p.Products)
			for {
				select {
				case <-stream.Context().Done():
					log.Println("..........ctx canceled...........", stream.Context().Err())
					return
				default:
					if len(p.Products) == 0 {
						p.FinishRequest <- 1
						return
					}
				}
			}
		}()
	}

	//output (work for from database and web)
	for {
		select {
		case product := <-p.Products:
			err := stream.Send(&product)
			if err != nil {
				log.Println("client closed")
			}
		case <-p.FinishRequest:
			log.Println("Done!")
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
