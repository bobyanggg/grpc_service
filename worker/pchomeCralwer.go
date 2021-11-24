package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/leopardx602/grpc_service/sql"
)

// func FindMaxPage(keyword string, errChan chan error) int {
//  // TODO: find maxpage
// }

type PChomeQuery struct {
	keyword string
}

type PchomeResponse struct {
	Prods []Commodity `json:"prods"`
}

type Commodity struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
	PicS  string `json:"picS"`
	Id    string `json:"Id"`
}

func NewPChomeQuery(keyword string) *PChomeQuery {
	return &PChomeQuery{
		keyword: keyword,
	}
}

func (q *PChomeQuery) Crawl(page int, finishQuery chan bool, newProducts chan *sql.Product, wgJob *sync.WaitGroup) {
	var client = &http.Client{Timeout: 10 * time.Second}

	//  minPrice := 10000
	//  maxPrice := 30000

	request, err := http.NewRequest("GET", "http://ecshweb.pchome.com.tw/search/v3.3/all/results?sort=rnk", nil)
	if err != nil {
		fmt.Println("Can not generate request")
		fmt.Println(err)
	}

	query := request.URL.Query()
	query.Add("q", q.keyword)
	//  query.Add("price", fmt.Sprintf("%d-%d", minPrice, maxPrice))

	var result PchomeResponse
	query.Set("page", fmt.Sprintf("%d", page))
	request.URL.RawQuery = query.Encode()
	url := request.URL.String()

	response, err := client.Get(url)
	if err != nil {
		log.Println(errors.Wrap(err, "can not get response form PChome"))
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		log.Println(errors.Wrapf(err, "can not decode JSON form PChome for %s", q.keyword))
	}

	defer response.Body.Close()

	for _, prod := range result.Prods {
		tempProduct := sql.Product{
			Word:       q.keyword,
			ProductID:  prod.Id,
			Name:       prod.Name,
			Price:      prod.Price,
			ImageURL:   "https://b.ecimg.tw" + prod.PicS,
			ProductURL: "https://24h.pchome.com.tw/prod/" + prod.Id,
		}
		newProducts <- &tempProduct
	}
	wgJob.Done()
}
