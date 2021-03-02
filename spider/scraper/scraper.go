package scraper

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// httpHeader es el tipo de dato que contieee las cabeceras de las peticiones http de los sitios web.
type httpHeaders map[string]string

// headers lista de httpHeader
type headers []httpHeaders

var headersPool = headers{
	{
		"DNT":             "1",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "es-US,es-419;q=0.9,es;q=0.8,en;q=0.7",
		"Cache-Control":   "max-age=0",
		"Connection":      "keep-alive",
	},
}

// Scraper es la clase para crear una instancia de la araña web
type Scraper struct {
	links        []string
	productNames []string
}

// New es el metodo que instancia la clase Scraper
func New() *Scraper {
	return &Scraper{}
}

// Links devuelve los enlaces obtenidos durante el raspado
func (c *Scraper) Links() []string { return c.links }

// ProductNames devuelve los nombres de los productos obtenidos
func (c *Scraper) ProductNames() []string { return c.productNames }

// GetAllUrls inicia el rasapado de datos
func (s *Scraper) GetAllUrls() {
	log.Println("Comenzando")
	c := colly.NewCollector(
		colly.AllowedDomains("https://www.mixup.com.mx", "www.mixup.com.mx", "mixup.com.mx"),
		colly.MaxDepth(5),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/69.0.3497.105 Mobile/15E148 Safari/605.1"),
	)

	extensions.Referer(c)

	/* rp, err := proxy.RoundRobinProxySwitcher("https://103.47.172.110:8080", "https://103.122.252.110:8080", "https://45.130.96.25:8080")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp) */

	log.Println("Collector creado")

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		//Delay:       120 * time.Second,
		RandomDelay: 300 * time.Second,
	})

	// callback
	c.OnRequest(func(r *colly.Request) {
		//r.Headers.Set("Referer", "https://facebook.com/")
		hds := headersPool[rand.Intn(len(headersPool))]
		for key, value := range hds {
			r.Headers.Set(key, value)
		}
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
