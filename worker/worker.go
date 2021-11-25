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
	inputChan chan *Job
	jobsChan  map[string]chan *Job
}

type Crawler interface {
	// Find product information from the website
	Crawl(page int, finishQuery chan bool, newProducts chan *sql.Product, wgJob *sync.WaitGroup)
}

// Start Consumer/worker and queue job
func Queue(ctx context.Context, keyWord string, newProducts chan *sql.Product) {
	c := NewConsumer()

	//responsible for start consumer, start worker
	go c.StartJob(ctx)

	maxPage := FindMaxPage(ctx, keyWord)
	fmt.Println("maxPage: ", maxPage)
	wgJobMomo := &sync.WaitGroup{}
	wgJobPchome := &sync.WaitGroup{}
	wgJobMomo.Add(maxPage)
	wgJobPchome.Add(maxPage)

	//momo
	go func(maxPage int) {
		for i := 1; i <= maxPage; i++ {
			input := &Job{i, "momo", keyWord, i, wgJobMomo, newProducts}
			fmt.Println("IN QUEUE", input)
			c.inputChan <- input
			log.Println("already send input value:", input)
		}
	}(maxPage)

	//pchome
	go func(maxPage int) {
		for i := 1; i <= maxPage; i++ {
			inputPcHome := &Job{i, "pchome", keyWord, i, wgJobPchome, newProducts}
			fmt.Println("IN QUEUE", inputPcHome)
			c.inputChan <- inputPcHome
			log.Println("already send input value:", inputPcHome)
		}
	}(maxPage)

	wgJobPchome.Wait()
	wgJobMomo.Wait()
	close(newProducts)

}

func (c Consumer) startConsumer(ctx context.Context) {
	for {
		select {
		case job := <-c.inputChan:
			if job.web == "momo" {
				c.jobsChan["momo"] <- job
			} else if job.web == "pchome" {
				c.jobsChan["pchome"] <- job
			}

			if ctx.Err() != nil {
				for _, jobchan := range c.jobsChan {
					close(jobchan)
				}
				return
			}
		case <-ctx.Done():
			for _, jobchan := range c.jobsChan {
				close(jobchan)
			}
			return
		}
	}
}

func (c *Consumer) process(num int, job Job, newProducts chan *sql.Product) {
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

func (c *Consumer) worker(ctx context.Context, num int, wg *sync.WaitGroup, web string) {
	defer wg.Done()
	log.Println("start the worker", num, web)
	for {
		select {
		case job := <-c.jobsChan[web]:
			if ctx.Err() != nil {
				log.Println("get next job", job, "and close the worker", num, web)
				return
			}
			c.process(num, *job, job.newProducts)
		case <-ctx.Done():
			log.Println("close the worker", num)
			return
		}
	}

}

const poolSize = 2

func (consumer *Consumer) StartJob(ctx context.Context) {
	webs := []string{"pchome", "momo"}
	fmt.Println("--------------start-------------")
	wg := &sync.WaitGroup{}

	//generate job channel for each web
	for _, val := range webs {
		consumer.jobsChan[val] = make(chan *Job, poolSize)
	}

	//generate workers for each web
	go func() {
		for _, web := range webs {
			for i := 0; i < poolSize; i++ {
				wg.Add(1)
				go consumer.worker(ctx, i, wg, web)
			}
		}
	}()

	go consumer.startConsumer(ctx)

}

func NewConsumer() *Consumer {
	return &Consumer{
		inputChan: make(chan *Job),
		jobsChan:  make(map[string]chan *Job),
	}
}
