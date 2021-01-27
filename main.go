package main

import (
	"fmt"
	"time"
)

func main() {
	/* initTimer := time.Now().UnixNano() / int64(time.Millisecond)
	spider := spider.New()
	//spider.StartCrawler()
	spider.StartScraper()
	endTimer := time.Now().UnixNano() / int64(time.Millisecond)
	fmt.Printf("Tiempo: %v s", (endTimer-initTimer)/1000) */
	testRotateIP()
}

func testRotateIP() {
	ipRotations := 3
	tor_handler := proxySwitcher.New()

	err := tor_handler.OpenURL()

	if err != nil {
		fmt.Println(err)
		return
	}

	for i = 0; i < ipRotations; i++ {
		err := tor_handler.ReNewConnection()
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(2 * time.Second)

		tor_handler.OpenURL()
	}
}
