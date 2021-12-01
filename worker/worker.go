package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	pb "gitlab.com/leopardx602/grpc_service/product"
	"gitlab.com/leopardx602/grpc_service/sql"
)

type WorkerConfig struct {
	MaxProduct int `json:"maxProduct"`
	WorkerNum  int `json:"workerNum"`
	SleepTime  int `json:"sleepTime"`
}

type Job struct {
	id          int
	web         string
	keyword     string
	page        int
	wgJob       *sync.WaitGroup
	newProducts chan *sql.Product
}

// Consumer struct
type Consumer struct {
	jobsChan map[string]chan *Job
}

type Crawler interface {
	// Find product information from the website
	Crawl(page int, finishQuery chan bool, newProducts chan *sql.Product, wgJob *sync.WaitGroup)
}

var webs = []string{"momo", "pchome"}

// Start Consumer/worker and queue job
func Queue(ctx context.Context, keyWord string, pProduct chan pb.UserResponse) {
	// load config
	jsonFile, err := os.Open("../config/worker.json")
	if err != nil {
		log.Fatal("faile to open json fail for creating worker: ", err)
	}
	log.Println("successfully opened worker config")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	var workerConfig WorkerConfig
	if err := json.NewDecoder(jsonFile).Decode(&workerConfig); err != nil {
		log.Fatal(err, "failed to decode worker config")
		return
	}

	totalWorker := workerConfig.WorkerNum
	sleepTime := workerConfig.SleepTime

	jobsChan := make(map[string]chan *Job)
	newProducts := make(chan *sql.Product, workerConfig.MaxProduct)

	//responsible for start consumer, start worker
	go startJob(ctx, jobsChan, totalWorker, sleepTime)

	go func() {
		for product := range newProducts {
			// Insert the data to the database.
			product.Word = keyWord
			if err := sql.Insert(*product); err != nil {
				log.Println(err)
			}

			// Push the data to grpc output.
			pProduct <- pb.UserResponse{
				Name:       product.Name,
				Price:      int32(product.Price),
				ImageURL:   product.ImageURL,
				ProductURL: product.ProductURL,
			}
		}
	}()

	for _, web := range webs {
		send(ctx, web, keyWord, newProducts, jobsChan)
	}

	close(newProducts)
}

func send(ctx context.Context, web, keyWord string, newProducts chan *sql.Product, jobsChan map[string]chan *Job) {
	var maxPage int

	// TODO : make a interface or merge existing?
	switch web {
	case "momo":
		maxPage = FindMaxMomoPage(ctx, keyWord)
	case "pchome":
		maxPage = FindMaxPchomePage(ctx, keyWord)
	}

	fmt.Println("maxPage: ", maxPage)
	wgJob := &sync.WaitGroup{}
	wgJob.Add(maxPage)

	go func(maxPage int) {
		for i := 1; i <= maxPage; i++ {
			input := &Job{i, web, keyWord, i, wgJob, newProducts}
			fmt.Println("IN QUEUE", input)
			jobsChan[web] <- input
			log.Println("already send input value:", input)
		}
	}(maxPage)

	wgJob.Wait()
}

func process(num int, job Job, newProducts chan *sql.Product, sleepTime int) {
	// n := getRandomTime()
	var crawler Crawler
	finishQuery := make(chan bool)
	log.Printf("%d starting on %v, Sleeping %d seconds...\n", num, job, sleepTime)

	switch job.web {
	case "momo":
		// crawler = NewMomoQuery(job.keyword)
		crawler = NewMomoQuery(job.keyword)
	case "pchome":
		crawler = NewPChomeQuery(job.keyword)
	}

	go crawler.Crawl(job.page, finishQuery, newProducts, job.wgJob)
	log.Println("finished", job.web, job.id)
	time.Sleep(time.Duration(sleepTime) * time.Second)

}

func worker(ctx context.Context, num int, wg *sync.WaitGroup, web string, jobsChan map[string]chan *Job, sleepTime int) {
	defer wg.Done()
	log.Println("start the worker", num, web)
	for {
		select {
		case job := <-jobsChan[web]:
			// stop worker if context error
			if ctx.Err() != nil && ctx.Err() != context.Canceled {
				log.Println("context err", ctx.Err())
				return
			}
			process(num, *job, job.newProducts, sleepTime)
		}
	}
}

func startJob(ctx context.Context, jobsChan map[string]chan *Job, totalWorker, sleepTime int) {
	fmt.Println("--------------start-------------")
	wg := &sync.WaitGroup{}

	//generate job channel for each web
	for _, val := range webs {
		jobsChan[val] = make(chan *Job, totalWorker)
	}

	//generate workers for each web
	go func() {
		for _, web := range webs {
			for i := 0; i < totalWorker; i++ {
				wg.Add(1)
				go worker(ctx, i, wg, web, jobsChan, sleepTime)
			}
		}
	}()
}
