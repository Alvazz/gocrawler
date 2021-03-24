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
	/* sw := scraper.NewSwitcher()
	for i := 0; i < 5; i++ {
		initConn := time.Now()
		sw.RotateIP()
		elapsedConn := time.Since(initConn)
		fmt.Println("Tiempo en conectar:", elapsedConn)
		time.Sleep(time.Duration(7 * time.Second))
	} */

	elapsed := time.Since(start)
	fmt.Println("Tiempo:", elapsed)
}
