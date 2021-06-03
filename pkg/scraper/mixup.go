package scraper

import (
	"context"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/internal/logging"
	"github.com/leosykes117/gocrawler/pkg/item"
)

type mixup struct {
	shop
}

// NewShopMixup crea un instancia d la estructura mixup
func newShopMixup(options ...func(*shop)) *mixup {
	m := &mixup{
		shop: shop{
			domainGlob:          `*mixup.*`,
			topLevelDomain:      "mixup.com",
			keywordsValue:       "Keywords",
			descriptionValue:    "Description",
			linkExtractionQuery: `(?m)https://www\.mixup\.com\.mx/[Mm]ixup/(([Hh]ome\.aspx)|((Categoria|Productos)\.aspx\?(etq\=))|(detproducto\.aspx\?sku=\d+)$)`,
			linkProductQuery:    `(?m)(https://www\.mixup\.com\.mx/[Mm]ixup/)(detproducto\.aspx\?sku=\d{12,})$`,
			allowedDomains: []string{
				"https://www.mixup.com.mx",
				"www.mixup.com.mx",
				"mixup.com.mx",
			},
		},
	}

	for _, f := range options {
		f(&m.shop)
	}
	return m
}

func (m *mixup) HTMLEvents(evts ...string) []OnHTMLEvent {
	events := make([]OnHTMLEvent, 0)
	for _, funcName := range evts {
		var e OnHTMLEvent
		switch funcName {
		case "GetMetaTags":
			e = m.GetMetaTags
		case "ExtractLinks":
			e = m.ExtractLinks
		case "GetProductDetails":
			e = m.GetProductDetails
		default:
			continue
		}
		events = append(events, e)
	}
	return events
}

// GetMetaTags obtiene el contenido de las etiquetas <meta> de la página web
func (m *mixup) GetMetaTags(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	return "html", func(e *colly.HTMLElement) {
		reqID := e.Request.Ctx.Get("ID")
		logging.InfoLogger.Println("Obteniedo las etiquetas meta[", reqID, "]")
		property := e.ChildAttr(`meta[property="og:image"]`, "content")
		imageURL := e.ChildAttr(`meta[name="twitter:image"]`, "content")
		keywords := e.ChildAttr(`meta[name="Description"]`, "content")
		description := e.ChildAttr(`meta[name="Keywords"]`, "content")
		logging.InfoLogger.Printf("[%s]Property: %s", reqID, property)
		logging.InfoLogger.Printf("[%s]Twitter: %s", reqID, imageURL)
		logging.InfoLogger.Printf("[%s]Keywords: %s", reqID, keywords)
		logging.InfoLogger.Printf("[%s]Description: %s", reqID, description)
		if onHTML != nil {
			onHTML(e)
		}
	}
}

func (m *mixup) ExtractLinks(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	return "a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "" {
			logging.WarningLogger.Println("No se encontro el link")
		} else {
			link = e.Request.AbsoluteURL(link)
			re := regexp.MustCompile(m.linkExtractionQuery)
			if !re.MatchString(link) {
				logging.WarningLogger.Printf("La URL no cumple las reglas para ser visitada: %s", link)
			}
			if onHTML != nil {
				onHTML(e)
			}
		}
	}
}

// GetProductDetails es el callback OnHTML de colly para obtener los detalles del producto
func (m *mixup) GetProductDetails(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	return "div.detail", func(e *colly.HTMLElement) {
		if strings.Contains(e.Request.URL.RawQuery, "sku=") {
			m.productDetails(e)
			if onHTML != nil {
				onHTML(e)
			}
		}
	}
}

// productDetails busca los detalles del producto en el página obtenida
func (m *mixup) productDetails(e *colly.HTMLElement) {
	var (
		detailCount                                int = 0
		name, brand, description, sourceStore, url string
		rating                                     float64
		reviews                                    item.Comments       = make(item.Comments, 0)
		details                                    item.ProductDetails = make(item.ProductDetails)
	)
	reqID := e.Request.Ctx.Get("ID")
	sourceStore = "Mixup"
	url = e.Request.AbsoluteURL(e.Request.URL.String())

	data := e.DOM.Text()
	spaceCleaner := regexp.MustCompile(`(?m)( {2,})`)
	divider := regexp.MustCompile(`(?m)(\r\n|\r|\n|\t)+`)
	data = spaceCleaner.ReplaceAllString(data, "")
	productData := divider.Split(data, -1)
	logging.InfoLogger.Printf("[%s]Detalles:\n%s", reqID, strings.Join(productData, "\n"))
	for _, info := range productData {
		info = strings.TrimSpace(info)
		if info != "" {
			detail := strings.Split(info, ":")
			switch detailCount {
			case 0:
				name = info
			case 1:
				brand = info
			default:
				if len(detail) > 1 {
					key := strings.TrimSpace(detail[0])
					value := strings.TrimSpace(detail[1])
					details[key] = value
				}
			}
			detailCount++
		}
	}

	description = e.DOM.Parent().NextAllFiltered("div.productcontent").Find("div#tabs-res").Text()
	description = strings.TrimSpace(description)
	productID, err := item.CreateID(name, sourceStore)
	if err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al crear el id del producto %q: %v", productID, err)
	}

	product := item.NewItem(
		item.ID(productID),
		item.Name(name),
		item.Brand(brand),
		item.Description(description),
		item.SourceStore(sourceStore),
		item.URL(url),
		item.Rating(rating),
		item.Reviews(reviews),
		item.Details(details),
	)

	if err := m.cacheService.CreateItem(context.Background(), product); err != nil {
		logging.ErrorLogger.Fatalf("Ocurrio un error al guardar el producto %s: %v", product.GetID(), err)
	}
}
