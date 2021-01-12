package spider

import (
	"fmt"
	"os"

	"github.com/leosykes117/gocrawler/spider/crawler"
	"github.com/leosykes117/gocrawler/spider/scraper"
)

type spider struct {
	c	*crawler.Crawler
	s	*scraper.Scraper
}

func New() *spider {
	return &spider {
		crawler.New(),
		scraper.New(),
	}
}

func (spd *spider) StartCrawler() {
	var seedUrl string = "https://www.coppel.com"
	fmt.Println("Vamos a comenzar")
	links, err := spd.c.GetAllUrls(seedUrl)
	if err != nil {
		fmt.Println("Ocurrio un error al extraer los links", err)
	}
	fmt.Println("Links extraidos")
	if err := saveUrls("./urls.txt", links); err != nil {
		fmt.Println("No se pudieron escribir los enlaces")
	}
	fmt.Println("Finalizado")
}

func (spd *spider) StartScraper() {
	fmt.Println("Vamos a comenzar")
	spd.s.GetAllUrls()
	
	fmt.Println("Escribiendo links")

	if err := saveUrls("./urls_from_colly.txt", spd.s.Links()); err != nil {
		fmt.Println("No se pudieron escribir los enlaces")
	}
	fmt.Println("Finalizado")
}

func saveUrls(filePath string, values []string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range values {
		fmt.Fprintln(f, value)
	}
	return nil
}