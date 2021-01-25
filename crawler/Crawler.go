package crawler

import (
	"container/list"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// NewCrawler returns a new Crawler instance.
func NewCrawler(siteURL string) *Crawler {

	siteURL = strings.Trim(siteURL, " ")
	siteURL = strings.TrimRight(siteURL, "/")

	var siteURLWWW string

	var baseURL string

	if strings.Index(siteURL, "http://") == 0 {
		siteURL = strings.Replace(siteURL, "https://", "", 1)
	} else if strings.Index(siteURL, "https://") == 0 {
		siteURL = strings.Replace(siteURL, "https://", "", 1)
	}

	fmt.Printf("URL %s \n", siteURL)

	if strings.Contains(siteURL, "www.") {
		siteURLWWW = siteURL
		siteURL = strings.Replace(siteURL, "www.", "", 1)
	} else {
		siteURLWWW = "www." + siteURL
	}

	baseURL = "https://" + siteURL + "/"

	fmt.Printf("Crawling %s\n", siteURL)

	return &Crawler{
		BaseURL:         baseURL,
		SiteURL:         siteURL,
		SiteURLWWW:      siteURLWWW,
		CrawlQueue:      list.New(),
		URLResponseList: make(map[string]int),
	}
}

// Crawler stores information about the websites crawl including
// a Count of how many pages have been visited, a queue of upcoming urls to visit
// and a list of responses
type Crawler struct {
	BaseURL         string
	SiteURLWWW      string
	SiteURL         string
	CrawlQueue      *list.List
	URLResponseList map[string]int
	CrawledCount    int
}

//CrawlAndScrape begins the crawling process, firstly it
//will crawl the SiteURL and then creates a queue of pages
//to crawl based off the urls found on the first page
func (c *Crawler) CrawlAndScrape() {

	var wg sync.WaitGroup

	//Crawl BaseURL to find the urls
	c.crawlURL(c.BaseURL)

	for {
		e := c.CrawlQueue.Front()

		if e == nil {
			break
		}

		c.crawlURL(e.Value.(string))

		c.CrawlQueue.Remove(e)

		fmt.Printf("URLS Queue: %d \n", c.CrawlQueue.Len())
		fmt.Printf("URLS Crawled: %d \n", c.CrawledCount)

		//Save progress every 20 crawls
		if c.CrawledCount%20 == 0 {
			go c.saveProgress(&wg)
			wg.Add(1)
		}
	}

	//Save progress one last time
	go c.saveProgress(&wg)
	wg.Add(1)

	wg.Wait()

}

func (c *Crawler) crawlURL(url string) {

	fmt.Printf("url: %s\n", url)

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	c.scrapeInternalUrls(resp.Body)

	c.URLResponseList[url] = resp.StatusCode

	c.CrawledCount++
}

func (c *Crawler) scrapeInternalUrls(data io.Reader) error {

	z := html.NewTokenizer(data)

	for {

		tt := z.Next()

		if tt == html.ErrorToken {
			return z.Err()
		}

		if name, _ := z.TagName(); string(name) != "a" {
			continue
		}

		var href string

		for {
			key, val, more := z.TagAttr()

			if string(key) == "href" {
				href = strings.Trim(string(val), " ")
			}

			if !more {
				break
			}
		}

		href = strings.Split(href, "#")[0]

		if href == "" {
			continue
		}

		if strings.Contains(href, "mailto:") {
			continue
		}

		if strings.Contains(href, "tel:") {
			continue
		}

		if strings.Index(href, "http") == 0 {

			if !strings.Contains(href, "://"+c.SiteURL) && !strings.Contains(href, "://"+c.SiteURLWWW) {
				continue
			}

			href = strings.Replace(href, "http://", "", 1)
			href = strings.Replace(href, "https://", "", 1)
			href = strings.Replace(href, c.SiteURL, "", 1)
			href = strings.Replace(href, c.SiteURLWWW, "", 1)
		}

		href = strings.TrimLeft(href, "/")
		href = c.BaseURL + href

		if _, exists := c.URLResponseList[href]; exists {
			continue
		}

		c.URLResponseList[href] = 0
		c.CrawlQueue.PushBack(href)
	}
}

func (c *Crawler) saveProgress(wg *sync.WaitGroup) {

	f, err := os.OpenFile(c.SiteURL+".csv", os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		panic(err)
	}

	w := csv.NewWriter(f)

	w.Write([]string{"url", "status"})

	for url, status := range c.URLResponseList {
		if status == 0 {
			continue
		}

		w.Write([]string{url, strconv.Itoa(status)})
	}

	w.Flush()

	wg.Done()
}
