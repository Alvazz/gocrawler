package main

import (
	"fmt"
	"time"

	"github.com/hako/durafmt"
	"github.com/leosykes117/gocrawler/scraper"
)

func main() {
	start := time.Now()
	spider := scraper.New()
	spider.GetAllUrls()
	elapsed := time.Since(start)
	fmt.Println("Tiempo:", durafmt.Parse(elapsed))
}
