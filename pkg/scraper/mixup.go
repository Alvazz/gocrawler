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

// GetMetaTags obtiene el contenido de las etiquetas <meta> de la página web
func (m *mixup) GetMetaTags(e *colly.HTMLElement) {
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
}

// GetProductDetails obtiene los datos del producto de la página
func (m *mixup) GetProductDetails(e *colly.HTMLElement) {
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
