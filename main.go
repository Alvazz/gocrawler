package main

import (
	"log"
	"time"

	"github.com/leosykes117/gocrawler/spider"
)

func main() {
	initTimer := time.Now().UnixNano() / int64(time.Millisecond)
	spider := spider.New()
	//spider.StartCrawler()
	spider.StartScraper()
	endTimer := time.Now().UnixNano() / int64(time.Millisecond)
	log.Printf("Tiempo: %v s", (endTimer-initTimer)/1000)
}
