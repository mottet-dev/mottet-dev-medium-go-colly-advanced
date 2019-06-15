package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/proxy"
	"github.com/mottet-dev/medium-go-colly-basics/utils"
)

func main() {
	fileName := "amazon_products.csv"
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Could not create %s", fileName)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Product Name", "Stars", "Price"})

	c := colly.NewCollector(
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		RandomDelay: 2 * time.Second,
		Parallelism: 4,
	})

	extensions.RandomUserAgent(c)

	proxySwitcher, err := proxy.RoundRobinProxySwitcher("socks5://188.226.141.127:1080", "socks5://67.205.132.241:1080")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(proxySwitcher)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
		fmt.Println("UserAgent", r.Headers.Get("User-Agent"))
	})

	c.OnHTML("div.s-result-list.s-search-results.sg-row", func(e *colly.HTMLElement) {
		e.ForEach("div.a-section.a-spacing-medium", func(_ int, e *colly.HTMLElement) {
			var productName, stars, price string

			productName = e.ChildText("span.a-size-medium.a-color-base.a-text-normal")

			if productName == "" {
				// If we can't get any name, we return and go directly to the next element
				return
			}

			stars = e.ChildText("span.a-icon-alt")
			utils.FormatStars(&stars)

			price = e.ChildText("span.a-price > span.a-offscreen")
			utils.FormatPrice(&price)

			writer.Write([]string{
				productName,
				stars,
				price,
			})
		})
	})

	for i := 1; i <= 20; i++ {
		fullURL := fmt.Sprintf("https://www.amazon.com/s?k=nintendo+switch&page=%d", i)
		c.Visit(fullURL)
	}
	c.Wait()
}
