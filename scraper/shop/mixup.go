package shop

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/logging"
)

type mixup struct {
	shop
	categoryLinkPattern string
	productLinkPatter   string
}

// NewShopMixup crea un instancia de la estructura mixup
func NewShopMixup() *mixup {
	return &mixup{
		categoryLinkPattern: `(?m)(Categoria|Productos).aspx\?(etq\=)`,
		shop: shop{
			topLevelDomain:      "mixup.com",
			keywordsValue:       "Keywords",
			descriptionValue:    "Description",
			linkExtractionQuery: ".itemlist.category a[href]",
		},
	}
}

// GetCategoryLinkPattern retorna el valor de mixup.categoryLinkPattern
func (m *mixup) GetCategoryLinkPattern() string {
	return m.categoryLinkPattern
}

// GetMetaTags obtiene el contenido de las etiquetas <meta> de la página web
func (m *mixup) GetMetaTags(e *colly.HTMLElement) {
	reqId := e.Request.Ctx.Get("ID")
	logging.InfoLogger.Println("Obteniedo las etiquetas meta[", reqId, "]")
	property := e.ChildAttr(`meta[property="og:image"]`, "content")
	twitter := e.ChildAttr(`meta[name="twitter:image"]`, "content")
	keywords := e.ChildAttr(`meta[name="Description"]`, "content")
	description := e.ChildAttr(`meta[name="Keywords"]`, "content")
	logging.InfoLogger.Printf("[%s]Property: %s", reqId, property)
	logging.InfoLogger.Printf("[%s]Twitter: %s", reqId, twitter)
	logging.InfoLogger.Printf("[%s]Keywords: %s", reqId, keywords)
	logging.InfoLogger.Printf("[%s]Description: %s", reqId, description)
}

// GetProductData obtiene los datos del producto y los almacena en un map
func (m *mixup) GetProductData(e *colly.HTMLElement) {
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
}

// GetProductDetails
func (m *mixup) GetProductDetails(e *colly.HTMLElement) {
	data := e.DOM.Text()
	spaceCleaner := regexp.MustCompile(`(?m)( {2,})`)
	divider := regexp.MustCompile(`(?m)(\r\n|\r|\n|\t)+`)
	data = spaceCleaner.ReplaceAllString(data, "")
	words := divider.Split(data, -1)
	productData := make([]string, 0)
	for _, w := range words {
		w := strings.TrimSpace(w)
		if w != "" {
			productData = append(productData, w)
		}
	}
	logging.InfoLogger.Printf("TEXTO DEL PRODUCTOS:\n%s", strings.Join(productData, "\n"))
}
