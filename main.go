package main

import (
	"fmt"
	"time"

	"github.com/leosykes117/gocrawler/scraper"
)

const (
	eqHours   uint64 = 3600000
	eqMinutes uint64 = 60000
	eqSeconds uint64 = 1000
)

func main() {
	initTimer := time.Now().UnixNano() / int64(time.Millisecond)

	/* spider := scraper.New()
	spider.GetAllUrls() */
	sw := scraper.NewSwitcher()
	for i := 0; i < 5; i++ {
		sw.RotateIP()
		time.Sleep(time.Duration(7 * time.Second))
	}

	endTimer := time.Now().UnixNano() / int64(time.Millisecond)
	fmt.Println("Tiempo")
	fmt.Println("\tHoras:", (endTimer-initTimer)/int64(eqHours))
	fmt.Println("\tMinutes:", (endTimer-initTimer)/int64(eqMinutes))
	fmt.Println("\tSegundos:", (endTimer-initTimer)/int64(eqSeconds))
}
