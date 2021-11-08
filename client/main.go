package main

import (
	"context"
	"io"
	"log"
	"os"

	pb "gitlab.com/leopardx602/grpc_service/product"

	"google.golang.org/grpc"
)

func main() {
	// get parameter
	if len(os.Args) < 2 {
		log.Fatal("no parameter")
	}
	searchKeyWord := os.Args[1]
	log.Println("Your input:", searchKeyWord)

	// connect to GRPC service
	conn, err := grpc.Dial(":8081", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	stream, err := client.GetUserInfo(context.Background(), &pb.UserRequest{KeyWord: searchKeyWord})
	if err != nil {
		log.Fatal(err)
	}

	// receive
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			log.Println("Done")
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("reply : %v\n", reply)
	}
}
