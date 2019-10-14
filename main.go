package main

import (
	"flag"
	"fmt"
	"net/url"

	"./crawler"
)

func main() {

	var siteURL string

	flag.StringVar(&siteURL, "url", "not_set", "Define URL of site to be crawled")

	flag.Parse()

	if siteURL == "not_set" {
		fmt.Println("-url flag not set.")
		return
	}

	url, err := url.Parse(siteURL)

	if err != nil {
		panic(err)
	}

	c := crawler.NewCrawler(url.String())
	c.CrawlAndScrape()

}
