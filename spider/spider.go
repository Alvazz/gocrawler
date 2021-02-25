package spider

import (
	"fmt"
	"os"
	"path/filepath"

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
	
	fmt.Println("Escribiendo los productos obtenidos...")

	filename, err := getFilePath("products.txt")
	fmt.Println("Ruta del archivo", filename)
	if err != nil {
		fmt.Println("Error al escribir el archivo")
		return
	}

	if err = saveUrls(filename, spd.s.ProductNames()); err != nil {
		fmt.Println("No se pudieron escribir los nombres de los productos")
		return
	}
	fmt.Println("Finalizado")
}

func getFilePath (filename string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	wd, _ := os.Getwd()
	fmt.Println("WD", wd)
	if err != nil {
			return "", err
	}	
	return dir + "/" + filename, nil
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