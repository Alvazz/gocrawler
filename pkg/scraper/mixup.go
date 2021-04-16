package scraper

import (
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/logging"
)

type mixup struct {
	shop
	categoryLinkPattern string
	productLinkPatter   string
}

// NewShopMixup crea un instancia de la estructura mixup
func newShopMixup() *mixup {
	return &mixup{
		shop: shop{
			topLevelDomain:      "mixup.com",
			keywordsValue:       "Keywords",
			descriptionValue:    "Description",
			linkExtractionQuery: `(?m)https://www\.mixup\.com\.mx/[Mm]ixup/(([Hh]ome\.aspx)|((Categoria|Productos)\.aspx\?(etq\=))|(detproducto\.aspx\?sku=\d{12})$)`,
			linkProductQuery:    `(?m)(https://www\.mixup\.com\.mx/[Mm]ixup/)(detproducto\.aspx\?sku=\d{12,})$`,
			//linkExtractionQuery: `(?m)(https://www\.mixup\.com\.mx/[Mm]ixup/)(Categoria|Productos)\.aspx\?(etq\=)`,
		},
	}
}

// GetCategoryLinkPattern retorna el valor de mixup.categoryLinkPattern
func (m *mixup) GetCategoryLinkPattern() string {
	return m.categoryLinkPattern
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
func (m *mixup) GetProductDetails(e *colly.HTMLElement, s *Scraper) {
	var (
		detailCount                                int = 0
		name, brand, description, sourceStore, url string
		rating                                     item.Score
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

	s.acquiredProducts = append(s.acquiredProducts, item.NewItem(
		name,
		brand,
		description,
		sourceStore,
		url,
		rating,
		reviews,
		details,
	))
}

/* func (m *mixup) GetProductData(e *colly.HTMLElement) {
	if !strings.Contains(e.Request.URL.RawQuery, "sku=") {
		return
	}
	reqID := e.Request.Ctx.Get("ID")
	productData := make(map[int][]string)

	deleteSpaces := func(s *string) {
		manySpace := regexp.MustCompile(`(?m)( {2,})`)
		*s = manySpace.ReplaceAllString(*s, " ")
		divider := regexp.MustCompile(`(?m)(\r\n|\r|\n)+`)
		*s = divider.ReplaceAllString(*s, "")
		*s = strings.TrimSpace(*s)
	}

	brandProduct := e.ChildText(`div[class*="titulo"]`)
	deleteSpaces(&brandProduct)
	logging.InfoLogger.Printf("[%s]Marca y Productos: %s", reqID, brandProduct)

	productDOM := e.DOM
	prevTag := ""
	tagID := 0
	productDOM.Contents().Each(func(i int, element *goquery.Selection) {
		nodeType := goquery.NodeName(element)
		if nodeType == "span" && element.HasClass("bold") {
			key := element.Text()
			deleteSpaces(&key)
			if key != "" {
				productData[tagID] = make([]string, 2)
				productData[tagID][0] = key
				prevTag = "span"
				logging.InfoLogger.Printf("%d[%d]#span: %s", i, tagID, element.Text())
			}
		} else if prevTag == "span" && nodeType == "#text" {
			logging.InfoLogger.Printf("[%d]#text: %s", tagID, element.Text())
			value := element.Text()
			deleteSpaces(&value)
			if value != "" {

				if _, ok := productData[tagID]; !ok {
					productData[tagID] = make([]string, 2)
				}
				productData[tagID][1] = value
				tagID++
			}
		} else {
			logging.InfoLogger.Printf("[%d]<%s>:%s</%s>", i, nodeType, element.Text(), nodeType)
		}
	})
	logging.InfoLogger.Printf("[%s]Descripción:\n\t%+v", reqID, productData)
} */
