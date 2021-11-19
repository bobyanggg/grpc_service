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

// func getRandomTime() int {
// 	rand.Seed(time.Now().UnixNano())
// 	return rand.Intn(10)
// }

// func withContextFunc(ctx context.Context, f func()) context.Context {
// 	ctx, cancel := context.WithCancel(ctx)
// 	go func() {
// 		c := make(chan os.Signal)
// 		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
// 		defer signal.Stop(c)
// 		select {
// 		case <-ctx.Done():
// 		case <-c:
// 			cancel()
// 			f()
// 		}
// 	}()

// 	return ctx
// }

func (c *Consumer) Queue(keyWord string, errChan chan error, newProducts chan *sql.Product, finishChan chan bool) {
	maxPage := FindMaxPage(keyWord, errChan)
	fmt.Println("maxPage: ", maxPage)
	wgJob := &sync.WaitGroup{}
	wgJob.Add(maxPage)
	go func(maxPage int) {

		for i := 1; i <= maxPage; i++ {
			input := &Job{i, "momo", keyWord, i, wgJob, newProducts}
			fmt.Println("IN QUEUE", input)
			c.inputChan <- input
			log.Println("already send input value:", input)
		}

	}(maxPage)

	wgJob.Wait()
	finishChan <- true

}

func (c Consumer) startConsumer(ctx context.Context) {
	for {
		select {
		case job := <-c.inputChan:
			if job.web == "momo" {
				c.jobsChan["momo"] <- job
			}
			if job.web == "pchome" {
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

func (c *Consumer) process(num int, job Job, newProducts chan *sql.Product, errChan chan error) {
	// n := getRandomTime()
	finishQuery := make(chan bool)
	n := 2
	log.Printf("%d starting on %v, Sleeping %d seconds...\n", num, job, n)

	go MomoCrawler(job.keyword, job.page, finishQuery, newProducts, errChan)
	time.Sleep(time.Duration(n) * time.Second)
	log.Println("finished", job.web, job.id)
	job.wgJob.Done()

}

func (c *Consumer) worker(ctx context.Context, num int, wg *sync.WaitGroup, web string, errChan chan error) {
	defer wg.Done()
	log.Println("start the worker", num, web)
	for {
		select {
		case job := <-c.jobsChan[web]:
			if ctx.Err() != nil {
				log.Println("get next job", job, "and close the worker", num, web)
				return
			}
			c.process(num, *job, job.newProducts, errChan)
		case <-ctx.Done():
			log.Println("close the worker", num)
			return
		}
	}

}

const poolSize = 2

func (consumer *Consumer) StartJob(ctx context.Context, errChan chan error) {
	//keyWord := "ipad"
	webs := []string{"pchome", "momo"}
	fmt.Println("--------------start-------------")
	// finished := make(chan bool)
	wg := &sync.WaitGroup{}

	//make consumer

	for _, val := range webs {
		consumer.jobsChan[val] = make(chan *Job, poolSize)
	}

	go func() {
		for _, web := range webs {
			for i := 0; i < poolSize; i++ {
				wg.Add(1)
				go consumer.worker(ctx, i, wg, web, errChan)
			}
		}
	}()

	go consumer.startConsumer(ctx)

	// go func() {
	// 	wgJobs.Wait()
	// 	fmt.Println("all finished")
	// 	close(finished)
	// }()

	// go func(maxPage int) {

	// 	for i := 1; i <= maxPage; i++ {
	// 		consumer.queue(&Job{i, "momo", keyWord, i}, wgJobs)
	// 	}

	// }(maxPage)

	// <-finished
	// finishChan <- true
	// log.Println("Crawl over")
}

func NewConsumer() *Consumer {
	return &Consumer{
		// inputChan: make(chan *Job),
		inputChan: make(chan *Job),
		jobsChan:  make(map[string]chan *Job),
	}
}
