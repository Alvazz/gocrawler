package scraper

import (
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/proxy"
	_ "github.com/gocolly/colly/proxy"
)

type Scraper struct {
	links        []string
	productNames []string
}

func New() *Scraper {
	return &Scraper{}
}

func (c *Scraper) Links() []string        { return c.links }
func (c *Scraper) ProductNames() []string { return c.productNames }

// Start inicia el crawler
func (s *Scraper) GetAllUrls() {
	log.Println("Comenzando")
	c := colly.NewCollector(
		colly.AllowedDomains("https://www.mixup.com.mx", "www.mixup.com.mx", "mixup.com.mx"),
		colly.MaxDepth(5),
		colly.Async(true),
	)

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	rp, err := proxy.RoundRobinProxySwitcher("socks5://127.0.0.1:9050", "socks5://127.0.0.1:9050", "http://161.35.4.201:80")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	log.Println("Collector creado")

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       5 * time.Second,
		RandomDelay: 20 * time.Second,
	})

	// callback
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://facebook.com/")
		r.Headers.Set("DNT", "1")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		log.Println("USER AGENT -->", r.Headers.Get("User-Agent"))

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
