package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	scrapy2 "github.com/crashiura/scrapy"
)

const categoryURL = "https://www.wildberries.ru/catalog/bytovaya-tehnika/krupnaya-bytovaya-tehnika"
const productURL = "https://www.wildberries.ru/catalog/11481158/detail.aspx?targetUrl=GP"

func main() {
	sc := scrapy2.NewScrapy()
	categoryHandler := getCategoryHandler()
	sc.AddHandler(categoryHandler)

	productHandler := getProductHandler()
	sc.AddHandler(productHandler)

	if err := productHandler.AddURL(context.Background(),productURL); err != nil {
		log.Fatal(err)
	}

	sc.Run()
	sc.Wait()
}

func getCategoryHandler() *scrapy2.Handler {
	categoryHandler := scrapy2.NewHandler()
	categoryHandler.Priority = 3

	categoryHandler.OnRequest(func(request *scrapy2.Request) {
		fmt.Println("Visit ", request.Request.URL)
	})

	categoryHandler.OnError(func(response *http.Response, err error) {
		fmt.Println("CODE: ", response.StatusCode)
		fmt.Println(err)
	})

	categoryHandler.OnHtml(func(request *scrapy2.Request, response *http.Response, doc *goquery.Document) {
		fmt.Println("CODE: ", response.StatusCode)

		//document.querySelector("#catalog-content > div.catalog_main_table.j-products-container > div:nth-child(1) > div > span > span > span > a")
		urls := make([]string, 0)
		doc.Find(".j-open-full-product-card").Each(func(i int, selection *goquery.Selection) {
			link, ok := selection.Attr("href")
			if ok {
				urls = append(urls, link)
			}
		})

		fmt.Println(urls)
	})

	return categoryHandler
}

func getProductHandler() *scrapy2.Handler {
	productHandler := scrapy2.NewHandler()
	productHandler.Priority = 0

	productHandler.OnRequest(func(request *scrapy2.Request) {
		fmt.Println("Visit ", request.Request.URL)
	})

	productHandler.OnError(func(response *http.Response, err error) {
		fmt.Println("CODE: ", response.StatusCode)
		fmt.Println(err)
	})

	productHandler.OnHtml(func(request *scrapy2.Request, response *http.Response, doc *goquery.Document) {
		fmt.Println("CODE: ", response.StatusCode)
	})

	return productHandler
}
