package main

import (
	"fmt"
	"time"

	"github.com/leosykes117/gocrawler/scraper"
)

func main() {
	start := time.Now()

	spider := scraper.New()
	spider.GetAllUrls()

	elapsed := time.Since(start)
	fmt.Println("Tiempo:", elapsed)
}
