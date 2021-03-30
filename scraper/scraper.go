package scraper

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/joho/godotenv"
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
	lock         *sync.RWMutex
	seedURL      string
	links        []string
	productNames []string
}

//env es un map que contiene las variable de ambiente del archivo .env
type enVars map[string]string

var crawlerVars enVars = make(enVars)
var envFilePath string

func init() {
	var err error
	projectPath, ok := os.LookupEnv("PROJECTPATH")
	if !ok {
		log.Fatalf("%s not set\n", "PROJECTPATH")
	} else {
		log.Printf("%s=%s\n", "PROJECTPATH", projectPath)
	}
	envFilePath, err = filepath.Abs(filepath.Join(projectPath, "./.env"))
	if err != nil {
		log.Fatalf("Error al obtener l ruta del archivo .env: %v", err)
	}
	crawlerVars, err = godotenv.Read(envFilePath)
	if err != nil {
		log.Fatalf("Error al leer .env: %v", err)
	}
	log.Println("Archivo .env leido correctamente")
}

// New es el metodo que instancia la clase Scraper
func New() *Scraper {
	log.Println(crawlerVars)
	return &Scraper{
		lock:    &sync.RWMutex{},
		seedURL: crawlerVars["SEEDURL"],
	}
}

// Links devuelve los enlaces obtenidos durante el raspado
func (s *Scraper) Links() []string { return s.links }

// ProductNames devuelve los nombres de los productos obtenidos
func (s *Scraper) ProductNames() []string { return s.productNames }

func (s *Scraper) setSeedURL(url string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.seedURL = url
}

// GetAllUrls inicia el rasapado de datos
func (s *Scraper) GetAllUrls() {
	log.Println("Comenzando")
	c := colly.NewCollector(
		colly.AllowedDomains("https://www.mixup.com.mx", "www.mixup.com.mx", "mixup.com.mx"),
		colly.MaxDepth(1),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 12_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/69.0.3497.105 Mobile/15E148 Safari/605.1"),
	)

	extensions.Referer(c)

	log.Println("Collector creado")

	c.SetRequestTimeout(140 * time.Second)
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay:       140 * time.Second,
		RandomDelay: 300 * time.Second,
	})

	// Se ejecuta antes de realizar la solicitud
	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visitando el sitio: %s\n", r.URL.String())
		hds := headersPool[rand.Intn(len(headersPool))]
		for key, value := range hds {
			r.Headers.Set(key, value)
		}
		reqID := randomString()
		r.Ctx.Put("RequestID", reqID)
	})

	// Se ejecuta si ocurre un error durante la petición
	c.OnError(func(r *colly.Response, e error) {
		log.Println("ERROR:", e, r.Request.URL, string(r.Body))
	})

	// Se ejecuta después de recibir la respuesta
	c.OnResponse(func(r *colly.Response) {
		log.Println("Página visitada", r.Request.URL)
		log.Printf("Cookies de la petición: %+v\n", r.Request.Headers.Get("Cookie"))
	})

	// Se ejecuta justo después de OnResponse si el contenido recibido es HTML
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "" {
			log.Println("No se encontro el link")
		} else {
			s.links = append(s.links, link)
			siteCookies := c.Cookies(link)
			if err := c.SetCookies(link, siteCookies); err != nil {
				log.Println("SET COOKIES ERROR: ", err)
			}
			c.Visit(e.Request.AbsoluteURL(link))
		}
	})

	c.OnHTML(".titulo", func(e *colly.HTMLElement) {
		title := strings.TrimSpace(e.Text)
		if title == "" {
			log.Println("No se encontró el título")
		} else {
			s.productNames = append(s.productNames, title)
		}
	})

	// Es el último callback en ejecutarse
	c.OnScraped(func(r *colly.Response) {
		s.setSeedURL(r.Request.URL.String())
	})

	// sitio inicial a visitar
	c.Visit(s.seedURL)
	c.Wait()
	log.Println("TERMINANDO EL SCRAPING")
	log.Println("Escribiendo los resultados")

	crawlerVars["SEEDURL"] = s.seedURL
	log.Printf(".env filepath: %s", envFilePath)
	err := godotenv.Write(crawlerVars, envFilePath)
	if err != nil {
		fmt.Printf("Error al escribir el archivo .env: %v", err)
	}

	filename, err := getFilePath("products.txt")
	if err != nil {
		log.Fatalf("Ocurrio un error al crear el archivo: %v", err)
	}
	err = s.saveUrls(filename)
	if err != nil {
		log.Fatalf("Ocurrio un error al escribir los elementos en el archivo: %v", err)
	}
	log.Printf("Archivo creado en %s\n", filename)
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
func (s *Scraper) saveUrls(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range s.productNames {
		fmt.Fprintln(f, value)
	}
	return nil
}

func randomString() string {
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
