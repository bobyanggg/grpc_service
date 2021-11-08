package model

import (
	"time"

	pb "gitlab.com/leopardx602/grpc_service/product"
)

type Product struct {
	Products      chan pb.UserResponse
	FinishRequest chan int
}

func (p *Product) AddProduct(product pb.UserResponse) {
	p.Products <- product
}

func (p *Product) Finish() {
	p.FinishRequest <- 1
}

func (p *Product) AddTest() {
	for i := 0; i < 5; i++ {
		p.AddProduct(pb.UserResponse{Name: "iphone12", Price: 26900, ImageURL: "no", ProductURL: "no"})
		time.Sleep(1 * time.Second)
	}
	p.Finish()
}
