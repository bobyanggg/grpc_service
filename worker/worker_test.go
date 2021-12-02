package worker

import (
	"context"
	"testing"

	pb "gitlab.com/leopardx602/grpc_service/product"
	"gitlab.com/leopardx602/grpc_service/sql"
)

func Test_Queue_Mouse(t *testing.T) {
	keyWord := "mouse"
	cleanupCtx := context.Background()
	pProduct := make(chan pb.UserResponse, 1000)
	results := []pb.UserResponse{}

	go func() {
		for product := range pProduct {
			results = append(results, product)
		}
	}()

	Queue(cleanupCtx, keyWord, pProduct)

	if len(results) == 0 {
		t.Error("error in Queue, result: ", results)
	}

}

func Test_sen_Mouse(t *testing.T) {
	ctx := context.Background()
	web, keyWord := "momo", "mouses"
	newProducts := make(chan *sql.Product, 200)
	jobsChan := make(map[string]chan *Job)
	results := []*Job{}
	workerConfig := WorkerConfig{WorkerNum: 2, MaxProduct: 200, SleepTime: 2}

	for _, val := range webs {
		jobsChan[val] = make(chan *Job, 2)
	}

	go func() {
		for {
			select {
			case job := <-jobsChan[web]:
				results = append(results, job)
			}
		}
	}()

	send(ctx, web, keyWord, newProducts, jobsChan, workerConfig)

	if len(results) == 0 {
		t.Error("error in Queue, result: ", results)
	}
}
