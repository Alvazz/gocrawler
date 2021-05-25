package scraper

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/leosykes117/gocrawler/internal/env"
	"github.com/leosykes117/gocrawler/internal/logging"
	"github.com/leosykes117/gocrawler/pkg/ciphersuite"
	"github.com/leosykes117/gocrawler/pkg/item"
)

func init() {
	logging.InitLogging()
	_, ok := os.LookupEnv("GO_CRAWLER_SEEDURL")
	if !ok {
		fmt.Println("Leyendo las variables del archivo")
		if err := env.LoadVars("go_crawler"); err != nil {
			logging.ErrorLogger.Fatal(err)
		}
	}
	if err := env.ReadVars("go_crawler"); err != nil {
		logging.ErrorLogger.Fatal(err)
	}
}

// Scraper es la clase para crear una instancia de la araña web
type Scraper struct {
	lock             *sync.RWMutex
	visitsCount      uint
	seedURL          string
	requests         scrapingRequests
	acquiredProducts item.Items
}

// New es el metodo que instancia la clase Scraper
func New() *Scraper {
	seedURL, _ := env.GetCrawlerVars(env.SeedURL)
	return &Scraper{
		lock:             &sync.RWMutex{},
		visitsCount:      0,
		seedURL:          seedURL.(string),
		requests:         make(scrapingRequests, 0),
		acquiredProducts: make(item.Items, 0),
	}
}

// setSeedURL utiliza un Mutex para guardar la última url visitada
func (s *Scraper) setSeedURL(url string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.seedURL = url
}

// addRequest utiliza Mutex para agregar una peticion al listado de peticiones.
func (s *Scraper) addRequest(rt *requestTracker) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.requests = append(s.requests, rt)
}

// GetAllUrls inicia el rasapado de datos
func (s *Scraper) GetAllUrls() {
	var shop shopCrawler = ShopFactory(Mixup)

	c := colly.NewCollector(
		colly.AllowedDomains(shop.GetAllowedDomains()...),
		//colly.MaxDepth(8),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0"),
		colly.URLFilters(
			regexp.MustCompile(shop.GetLinkExtractionQuery()),
		),
	)

	extensions.Referer(c)

	c.SetRequestTimeout(30 * time.Second)

	c.WithTransport(&http.Transport{
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 30 * time.Second,
	})

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  shop.GetDomainGlob(),
		Parallelism: 4,
		RandomDelay: 6 * time.Second,
	})
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al establecer los limites para el colector: %v", err)
	}

	// Se ejecuta antes de realizar la solicitud
	c.OnRequest(func(r *colly.Request) {
		reqID, _ := ciphersuite.GetMD5Hash(r.URL.String())
		logging.InfoLogger.Printf("[%s]Visitando el sitio: %s\n", reqID, r.URL.String())
		hds := GetHeaders()
		for key, value := range hds {
			r.Headers.Set(key, value)
		}
		r.Ctx.Put("ID", reqID)
		r.Ctx.Put("StartAt", time.Now().Format(time.UnixDate))
	})

	// Se ejecuta si ocurre un error durante la petición
	c.OnError(func(r *colly.Response, e error) {
		reqID := r.Ctx.Get("ID")
		strStartAt := r.Ctx.Get("StartAt")
		timeStartAt, err := time.Parse(time.UnixDate, strStartAt)
		if err != nil {
			logging.WarningLogger.Printf("Error al parsear la fecha: %v", err)
		}
		logging.ErrorLogger.Printf("OnError:%s\n\tID: %s,\n\tStartAt: %s", e, r.Ctx.Get("ID"), strStartAt)

		debugReq, err := env.GetCrawlerVars(env.DebugRequests)
		if err != nil {
			logging.ErrorLogger.Printf("Error la obtener la bandera de debug")
		}

		if ok := debugReq.(bool); ok {
			rt := newRequestTracker(
				reqID,
				r.Request.AbsoluteURL(r.Request.URL.String()),
				"OnError",
				r.Request,
				r,
				timeStartAt,
				time.Now(),
				e,
			)
			s.addRequest(rt)
		}
	})

	// Se ejecuta después de recibir la respuesta
	c.OnResponse(func(r *colly.Response) {
		re := regexp.MustCompile(shop.GetLinkExtractionQuery())
		url := r.Request.URL.String()
		if !re.MatchString(url) && !strings.Contains(url, "?sku=") {
			logging.WarningLogger.Printf("La URL no cumple las reglas para ser visitada: %s", url)
			return
		}
		reqID := r.Ctx.Get("ID")
		strStartAt := r.Ctx.Get("StartAt")
		timeStartAt, err := time.Parse(time.UnixDate, strStartAt)
		if err != nil {
			logging.WarningLogger.Printf("Error al parsear la fecha: %v", err)
		}
		debugReq, err := env.GetCrawlerVars(env.DebugRequests)
		if err != nil {
			logging.ErrorLogger.Printf("Error la obtener la bandera de debug")
		}

		if ok := debugReq.(bool); ok {
			rt := newRequestTracker(
				reqID,
				r.Request.AbsoluteURL(r.Request.URL.String()),
				"OnResponse",
				r.Request,
				r,
				timeStartAt,
				time.Now(),
				nil,
			)
			s.addRequest(rt)
		}
		logging.InfoLogger.Printf("OnResponse:\n\tID: %s,\nStartAt: %s", r.Ctx.Get("ID"), strStartAt)
	})

	// Se ejecuta justo después de OnResponse si el contenido recibido es HTML
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "" {
			logging.WarningLogger.Println("No se encontro el link")
		} else {
			link = e.Request.AbsoluteURL(link)
			re := regexp.MustCompile(shop.GetLinkExtractionQuery())
			if !re.MatchString(link) {
				logging.WarningLogger.Printf("La URL no cumple las reglas para ser visitada: %s", link)
			}
			siteCookies := c.Cookies(link)
			if err := c.SetCookies(link, siteCookies); err != nil {
				logging.ErrorLogger.Println("SET COOKIES ERROR: ", err)
			}
			s.visitsCount++
			err := c.Visit(link)
			if err != nil {
				logging.ErrorLogger.Printf("[%s][%s]Ocurrio un error al crear la petición: %v", e.Request.Ctx.Get("ID"), e.Request.AbsoluteURL(link), err)
			}
		}
	})

	c.OnHTML("html", shop.GetMetaTags)
	c.OnHTML("div.detail", func(e *colly.HTMLElement) {
		if strings.Contains(e.Request.URL.RawQuery, "sku=") {
			shop.GetProductDetails(e)
		}
	})

	// Es el último callback en ejecutarse
	c.OnScraped(func(r *colly.Response) {
		s.setSeedURL(r.Request.URL.String())
	})

	// sitio inicial a visitar
	s.visitsCount++
	err = c.Visit(s.seedURL)
	if err != nil {
		logging.ErrorLogger.Printf("Ocurrio un error al crear la petición de la URL semilla: %v", err)
	}
	c.Wait()
	logging.InfoLogger.Println("Escribiendo los resultados")

	err = env.SetCrawlerVars(env.SeedURL, s.seedURL)
	if err != nil {
		fmt.Printf("Error al escribir la última URL visitada: %v", err)
	}

	err = env.WriteVars("go_crawler")
	if err != nil {
		fmt.Printf("Error al escribir el archivo .env: %v", err)
	}

	requestsJSON, err := s.requests.MarshalJSON()
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al hacer Marshal de las solicitudes:\n%v", err)
	}
	jsonFile, err := getFilePath("scraping_request.json")
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al crear la ruta del archivo json: %v", err)
	}
	err = ioutil.WriteFile(jsonFile, requestsJSON, 0600)
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al crear el archivo json: %v", err)
	}
}

// getFilePath Crear la ruta del donde escribir el archivo
func getFilePath(filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fmt.Println(home)
	dir, err := filepath.Abs(filepath.Join(home, "./crawling-data/outs/"))
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(err)
			fmt.Println("Creando dir")
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return filepath.Join(dir, filename), nil
}

// randomString genera una cadena de caracteres aleatorios
func _() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 12
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
