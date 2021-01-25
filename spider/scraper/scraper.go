package scraper

import (
	"strings"
	"time"
	"log"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
)

type Scraper struct {
	links			[]string
	productNames	[]string
}

func New() *Scraper {
	return &Scraper {}
}

func (c *Scraper) Links() []string { return c.links }
func (c *Scraper) ProductNames() []string { return c.productNames }

// Start inicia el crawler
func (s *Scraper) GetAllUrls() {
	log.Println("Comenzando")
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
		colly.AllowedDomains("https://www.mixup.com.mx", "www.mixup.com.mx", "mixup.com.mx"),
		colly.MaxDepth(5),
		colly.Async(true),
	)

	rp, err := proxy.RoundRobinProxySwitcher("http://187.243.253.2:8080", "http://162.144.106.245:3838", "http://161.35.4.201:80")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	log.Println("Collector creado")

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay: 45 * time.Second,
		//RandomDelay: 45 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		''
		r.Headers.Set("Referer", "https://www.facebook.com/")
	})

	// callback,para saber que pagina se ha visitado
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "" {
			log.Println("No se encontro el link")
		} else {
			s.links = append(s.links, link)
			e.Request.Visit(link)
		}
	})

	c.OnHTML(".titulo", func(e *colly.HTMLElement) {
		log.Println("TITULO OBTENIDO")
		log.Printf("element %+v\n", e)
		title := strings.TrimSpace(e.Text)
		if title == "" {
			log.Println("No se encontró el título")
		} else {
			s.productNames = append(s.productNames, title)
		}
	})

	// callback
	c.OnError(func(r *colly.Response, e error) {
		log.Println("error:", e, r.Request.URL, string(r.Body))
	})

	// callback, nos reponde con la pagina que esta visitando
	c.OnResponse(func(r *colly.Response) {
		log.Println("Visitado", r.Request.URL)
	})

	// sitio que vamos a visitar
	c.Visit("https://www.mixup.com.mx")
	log.Println("Despues de visit")
	c.Wait()
	log.Println("TERMINANDO EL SCRAPING")
}