package scraper

import (
	"time"
	"log"

	"github.com/gocolly/colly"
	_"github.com/gocolly/colly/proxy"
)

type Scraper struct {
	links []string
}

func New() *Scraper {
	return &Scraper {}
}

func (c *Scraper) Links() []string { return c.links }

// Start inicia el crawler
func (s *Scraper) GetAllUrls() {
	log.Println("Comenzando")
	c := colly.NewCollector(
		//colly.AllowedDomains("https://www.coppel.com", "https://www.coppel.com/"),
		//colly.Debugger(&debug.LogDebugger{}),
		colly.MaxDepth(2),
		colly.Async(true),
		
	)

	log.Println("Collector creado")

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay: 5 * time.Second,
		RandomDelay: 15 * time.Second,
	})

	// callback,para saber que pagina se ha visitado
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Println("LINK -->", link)
		if link == "" {
			log.Println("No se encontro el link")
		} else {
			s.links = append(s.links, link)
			e.Request.Visit(link)
		}
	})

	// callback, nos reponde con la pagina que esta visitando
	c.OnResponse(func(r *colly.Response) {
		log.Println("Visitado", r.Request.URL)
	})

	// sitio que vamos a visitar
	c.Visit("https://www.coppel.com")
	log.Println("Despues de visit")
	c.Wait()
}