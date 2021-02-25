package crawler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {}

func New() *Crawler {
	return &Crawler{}
}


func (c *Crawler) GetAllUrls(seedURL string) (links []string, err error) {
	fmt.Println("Iniciando la extracción de links")
	client := &http.Client {
		Timeout: 30 * time.Second,
	}
	request, err := http.NewRequest("GET", seedURL, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Configurando los header de la petición 
	// Evitar que se cacheen las respuestas
	request.Header.Set("pragma", "no-cache")
	request.Header.Set("cache-control", "no-cache")
	// No ser rastreado en el sitio de destino.
	request.Header.Set("dnt", "1")
	// Preferencia del cliente por una respuesta encriptada y autenticada
	request.Header.Set("upgrade-insecure-requests", "1")
	// Dirección de la página precvia desde la cual un link no ha redirijido
	request.Header.Set("referer", "https://www.shopify.com.mx/")

	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		doc, e := goquery.NewDocumentFromReader(resp.Body)
		if e != nil {
			err = e
			fmt.Println(err)
			return
		}
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")
			links = append(links, link)
		})
	}
	return
}