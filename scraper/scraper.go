package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/joho/godotenv"
	"github.com/leosykes117/gocrawler/logging"
	"github.com/segmentio/ksuid"
)

var (
	headersPool = headers{
		{
			"DNT":             "1",
			"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"Accept-Encoding": "gzip, deflate, br",
			"Accept-Language": "es-US,es-419;q=0.9,es;q=0.8,en;q=0.7",
			"Cache-Control":   "max-age=0",
			"Connection":      "keep-alive",
		},
	}
	crawlerVars enVars = make(enVars)
	envFilePath string
)

type products []*Item

// Scraper es la clase para crear una instancia de la araña web
type Scraper struct {
	lock             *sync.RWMutex
	visitsCount      uint
	seedURL          string
	requests         scrapingRequests
	acquiredProducts products
}

func init() {
	logging.InitLogging()
	var err error
	projectPath, ok := os.LookupEnv("PROJECTPATH")
	if !ok {
		logging.ErrorLogger.Fatalf("%s not set\n", "PROJECTPATH")
	} else {
		logging.InfoLogger.Printf("%s=%s\n", "PROJECTPATH", projectPath)
	}
	envFilePath, err = filepath.Abs(filepath.Join(projectPath, "./.env"))
	if err != nil {
		logging.ErrorLogger.Fatalf("Error al obtener l ruta del archivo .env: %v", err)
	}
	crawlerVars, err = godotenv.Read(envFilePath)
	if err != nil {
		logging.ErrorLogger.Fatalf("Error al leer .env: %v", err)
	}
	logging.InfoLogger.Println("Archivo .env leido correctamente")
}

// New es el metodo que instancia la clase Scraper
func New() *Scraper {
	logging.InfoLogger.Println(crawlerVars)
	return &Scraper{
		lock:             &sync.RWMutex{},
		visitsCount:      0,
		seedURL:          crawlerVars["SEEDURL"],
		requests:         make(scrapingRequests, 0),
		acquiredProducts: make(products, 0),
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
	var shop shopCrawler = newShopMixup()

	c := colly.NewCollector(
		colly.AllowedDomains("https://www.mixup.com.mx", "www.mixup.com.mx", "mixup.com.mx"),
		//colly.MaxDepth(8),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0"),
	)

	extensions.Referer(c)

	c.SetRequestTimeout(100 * time.Second)
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       90 * time.Second,
		RandomDelay: 110 * time.Second,
	})

	// Se ejecuta antes de realizar la solicitud
	c.OnRequest(func(r *colly.Request) {
		reqID, _ := ksuid.NewRandom()
		logging.InfoLogger.Printf("[%s]Visitando el sitio: %s\n", reqID.String(), r.URL.String())
		hds := headersPool[rand.Intn(len(headersPool))]
		for key, value := range hds {
			r.Headers.Set(key, value)
		}
		r.Ctx.Put("ID", reqID.String())
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
	})

	// Se ejecuta después de recibir la respuesta
	c.OnResponse(func(r *colly.Response) {
		reqID := r.Ctx.Get("ID")
		strStartAt := r.Ctx.Get("StartAt")
		timeStartAt, err := time.Parse(time.UnixDate, strStartAt)
		if err != nil {
			logging.WarningLogger.Printf("Error al parsear la fecha: %v", err)
		}
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
		//log.Println("Página visitada", r.Request.URL)
		logging.InfoLogger.Printf("OnResponse:\n\tID: %s,\nStartAt: %s", r.Ctx.Get("ID"), strStartAt)
	})

	c.OnResponse(func(r *colly.Response) {
		logging.InfoLogger.Println("Verificando si el response es una imagen")
		if strings.Index(r.Headers.Get("Content-Type"), "image") > -1 {
			logging.InfoLogger.Printf("Imagen obtenida[%s]", r.Request.Ctx.Get("ID"))
		}
	})

	// Se ejecuta justo después de OnResponse si el contenido recibido es HTML
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "" {
			logging.WarningLogger.Println("No se encontro el link")
		} else {
			re := regexp.MustCompile(shop.GetLinkExtractionQuery())
			if !re.MatchString(link) && !strings.Contains(link, "?sku=") {
				logging.WarningLogger.Printf("La URL no cumple las reglas para ser visitada: %s", link)
			} else {
				siteCookies := c.Cookies(link)
				if err := c.SetCookies(link, siteCookies); err != nil {
					logging.ErrorLogger.Println("SET COOKIES ERROR: ", err)
				}
				s.visitsCount++
				c.Visit(e.Request.AbsoluteURL(link))
			}
		}
	})

	c.OnHTML("html", shop.GetMetaTags)
	c.OnHTML("div.detail", func(e *colly.HTMLElement) {
		if strings.Contains(e.Request.URL.RawQuery, "sku=") {
			shop.GetProductDetails(e, s)
		}
	})

	// Es el último callback en ejecutarse
	c.OnScraped(func(r *colly.Response) {
		s.setSeedURL(r.Request.URL.String())
	})

	// sitio inicial a visitar
	s.visitsCount++
	c.Visit(s.seedURL)
	c.Wait()
	logging.InfoLogger.Println("Escribiendo los resultados")

	crawlerVars["SEEDURL"] = s.seedURL
	logging.InfoLogger.Printf(".env filepath: %s", envFilePath)
	err := godotenv.Write(crawlerVars, envFilePath)
	if err != nil {
		fmt.Printf("Error al escribir el archivo .env: %v", err)
	}

	filename, err := getFilePath("products.json")
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al crear el archivo: %v", err)
	}
	err = s.saveProducts(filename)
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al escribir los elementos en el archivo: %v", err)
	}
	logging.InfoLogger.Printf("Archivo creado en %s\n", filename)

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

// saveUrls escribe en un archivo los productos obtenidos
func (s *Scraper) saveProducts(filePath string) error {
	if len(s.acquiredProducts) > 0 {
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		jsonProducts, err := json.MarshalIndent(s.acquiredProducts, "", "\t")
		if err != nil {
			return err
		}
		_, err = f.Write(jsonProducts)
		if err != nil {
			return err
		}
	}
	return nil
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
