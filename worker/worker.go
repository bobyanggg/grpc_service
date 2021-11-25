package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gitlab.com/leopardx602/grpc_service/sql"
)

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
func Queue(ctx context.Context, keyWord string, newProducts chan *sql.Product) {
	jobsChan := make(map[string]chan *Job)

	//responsible for start consumer, start worker
	go startJob(ctx, jobsChan)

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

func process(num int, job Job, newProducts chan *sql.Product) {
	// n := getRandomTime()
	var crawler Crawler
	finishQuery := make(chan bool)
	n := 2
	log.Printf("%d starting on %v, Sleeping %d seconds...\n", num, job, n)

	switch job.web {
	case "momo":
		crawler = NewMomoQuery(job.keyword)
	case "pchome":
		crawler = NewPChomeQuery(job.keyword)
	}

	go crawler.Crawl(job.page, finishQuery, newProducts, job.wgJob)
	time.Sleep(time.Duration(n) * time.Second)
	log.Println("finished", job.web, job.id)

}

func worker(ctx context.Context, num int, wg *sync.WaitGroup, web string, jobsChan map[string]chan *Job) {
	defer wg.Done()
	log.Println("start the worker", num, web)
	for {
		select {
		case job := <-jobsChan[web]:
			if ctx.Err() != nil {
				log.Println("get next job", job, "and close the worker", num, web)
				return
			}
			process(num, *job, job.newProducts)
		case <-ctx.Done():
			log.Println("close the worker", num)
			return
		}
	}

}

const poolSize = 2

func startJob(ctx context.Context, jobsChan map[string]chan *Job) {
	fmt.Println("--------------start-------------")
	wg := &sync.WaitGroup{}

	//generate job channel for each web
	for _, val := range webs {
		jobsChan[val] = make(chan *Job, poolSize)
	}

	//generate workers for each web
	go func() {
		for _, web := range webs {
			for i := 0; i < poolSize; i++ {
				wg.Add(1)
				go worker(ctx, i, wg, web, jobsChan)
			}
		}
	}()
}
