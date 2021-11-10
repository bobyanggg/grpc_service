package findtest

import (
	"context"
	"strconv"
	"time"

	"gitlab.com/leopardx602/grpc_service/sql"
)

func FindInWeb(ctx context.Context, products chan sql.Product, keyword string) {
	for i := 0; i < 5; i++ {
		productID := strconv.FormatInt(time.Now().Unix(), 10)
		products <- sql.Product{ProductID: productID, Name: "iphone12", Price: 26900, ImageURL: "no", ProductURL: "no"}
		time.Sleep(1 * time.Second)
	}
	close(products)
}
