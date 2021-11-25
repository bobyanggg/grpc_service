package worker

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"gitlab.com/leopardx602/grpc_service/sql"
)

func GetHttpHtmlContent(url string, selector string, sel interface{}) (string, error) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true), // debug using
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	}
	//Initialization parameters, first pass an empty data
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	c, _ := chromedp.NewExecAllocator(context.Background(), options...)

	// create context
	chromeCtx, _ := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	//Execute an empty task to create a chrome instance in advance
	chromedp.Run(chromeCtx, make([]chromedp.Action, 0, 1)...)

	//Create a context with a timeout of 40s
	timeoutCtx, cancel := context.WithTimeout(chromeCtx, 40*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector),
		chromedp.OuterHTML(sel, &htmlContent, chromedp.ByJSPath),
	)
	if err != nil {
		fmt.Printf("Run err : %v\n", err)
		return "", err
	}
	//log.Println(htmlContent)

	return htmlContent, nil
}

//Get specific data
func GetSpecialData(htmlContent string, selector string) (string, error) {
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Printf("Go Query error: %v", err)
		return "", err
	}

	var str string
	dom.Find(selector).Each(func(i int, selection *goquery.Selection) {
		str = selection.Text()
	})
	return str, nil
}

func FindMaxPage(ctx context.Context, keyword string) int {
	totalPageResult := 0
	starturl := fmt.Sprintf("https://www.momoshop.com.tw/search/searchShop.jsp?keyword=%s&searchType=1&curPage=%d", keyword, 1)
	selector := "#BodyBase > div.bt_2_layout.searchbox.searchListArea.selectedtop > div.searchPrdListArea.bookList > div.listArea > ul " //
	sel := `document.querySelector("body")`
	fmt.Println("getting maximum page @", starturl)
	html, err := GetHttpHtmlContent(starturl, selector, sel)
	if err != nil {
		log.Println(errors.Wrap(err, "failed to get total page"))
	}

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Println(errors.Wrap(err, "failed to go query"))
	}

	dom.Find("#BodyBase > div.bt_2_layout.searchbox.searchListArea.selectedtop > div.pageArea.topPage > dl > dt > span:nth-child(2)").Each(func(i int, selection *goquery.Selection) {
		pageStr := strings.Split(selection.Text(), "/")
		totalPage, _ := strconv.Atoi(pageStr[1])
		//fmt.Println("TotalPage.................................:", totalPage)
		totalPageResult = totalPage
	})
	return totalPageResult
}

type MomoQuery struct {
	keyword string
}

func NewMomoQuery(keyword string) *MomoQuery {
	return &MomoQuery{
		keyword: keyword,
	}
}

func (q *MomoQuery) Crawl(page int, finishQuery chan bool, newProducts chan *sql.Product, wgJob *sync.WaitGroup) {
	starturl := fmt.Sprintf("https://www.momoshop.com.tw/search/searchShop.jsp?keyword=%s&searchType=1&curPage=%d", q.keyword, page)
	absoluteURL := "https://www.momoshop.com.tw/"
	selector := "#BodyBase > div.bt_2_layout.searchbox.searchListArea.selectedtop > div.searchPrdListArea.bookList > div.listArea > ul " //
	sel := `document.querySelector("body")`
	fmt.Println("visiting url.....", starturl)
	html, err := GetHttpHtmlContent(starturl, selector, sel)
	if err != nil {
		log.Println(errors.Wrapf(err, "failed to visit url: %s", starturl))
	}

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Println(errors.Wrapf(err, "failed to create goquery"))
	}

	dom.Find("#BodyBase > div.bt_2_layout.searchbox.searchListArea.selectedtop > div.searchPrdListArea.bookList > div.listArea > ul > li").Each(func(i int, selection *goquery.Selection) {
		tempProduct := sql.Product{}

		tempProduct.Name = selection.Find("a > div.prdInfoWrap > h3.prdName").Text()
		if tempProduct.Name == "" {
			log.Println(errors.Wrapf(err, "failed to get price of: %s", tempProduct.Name))
		}

		tempPrice := strings.ReplaceAll(selection.Find("a > div.prdInfoWrap > p.money > span.price > b").Text(), ",", "")
		tempProduct.Price, err = strconv.Atoi(tempPrice)
		if err != nil {
			log.Println(errors.Wrapf(err, "failed to get price of: %s", tempProduct.Name))
		}

		tempImage, isFound := selection.Find("img.prdImg.lazy.lazy-loaded").Attr("src")
		if !isFound {
			(errors.Wrapf(err, "failed to find Image of %s", tempProduct.Name))
		}
		tempProduct.ImageURL = tempImage

		tempUrl, isFound := selection.Find("a.goodsUrl").Attr("href")
		if !isFound {
			log.Println(errors.Wrapf(err, "failed to find Url of %s", tempProduct.Name))
		}
		tempProduct.ProductURL = absoluteURL + tempUrl

		query, err := url.Parse(tempProduct.ProductURL)
		if err != nil {
			log.Println(errors.Wrapf(err, "failed to find Product Url of %s", tempProduct.Name))
		}
		querys := query.Query()
		tempProduct.ProductID = querys["i_code"][0]
		if tempProduct.ProductID == "" {
			log.Println(errors.Wrapf(err, "failed to find Product Url of %s", tempProduct.Name))
		}

		tempProduct.Word = q.keyword
		newProducts <- &tempProduct

	})
	wgJob.Done()
}
