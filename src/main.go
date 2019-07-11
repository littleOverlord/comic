package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	for i := 1; i <= 1074; i++ {
		go func(p int) {
			find(p)
		}(i)
	}
}

func find(page int) {
	// Request the HTML page.
	res, err := http.Get(fmt.Sprintf("http://www.2nunu.com/index.php?s=%2Findex-html&page=%d", page))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find(".in_01 li").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		a := s.Find("a").First()
		url, ok := a.Attr("href")
		if !ok {
			fmt.Println("don't find href")
		}
		img, ok := s.Find("img").Attr("src")
		if !ok {
			fmt.Println("don't find src")
		}
		title := a.Text()
		fmt.Printf("Review %d: %s - %s - %s\n", i, title, url, img)
	})
}
