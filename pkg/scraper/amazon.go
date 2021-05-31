package scraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/pkg/item"
)

// amazon es la estructura que implmenta la interfaz shopCrawler para scrapear
// la tienda amazon.com
type amazon struct {
	shop
}

// query: div.s-main-slot.s-result-list.s-search-results.sg-row a
// query data product: div#centerCol
// title: span.product-title-word-break
// brand: div#bylineInfo_feature_div a
// price: div#unifiedPrice_feature_div[data-feature-name="unifiedPrice"][data-cel-widget="unifiedPrice_feature_div"] span#priceblock_ourprice

// https://www.amazon.com.mx/Gildan-Camiseta-Interior-Hombres-Charcoal/dp/B077ZKK9YB/ref=sr_1_1?dchild=1&keywords=ropa&qid=1622402633&sr=8-1
// https://www.amazon.com.mx/Apple-Nuevo-MacBook-Chip-Pulgadas/dp/B08N6ST99B/ref=sr_1_2_sspa?dchild=1&keywords=macbook&qid=1622421746&sr=8-2-spons&th=1
// https://www.amazon.com.mx/DEKITA-Lavandería-Organizador-Compartimentos-Almacenamiento/dp/B08GZ1DXRG/ref=sr_1_2?dchild=1&keywords=ropa&qid=1622420598&sr=8-2

// newShopAmazon crea un instancia d la estructura amazon
func newShopAmazon(options ...func(*shop)) *amazon {
	a := &amazon{
		shop: shop{
			domainGlob:          `*amazon.*`,
			linkExtractionQuery: `(?m)https:\/\/www\.amazon\.(com\.mx|mx|es|co\.uk|com)\/(s[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$)|([\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+dp[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+)`,
			linkProductQuery:    `(?m)https:\/\/www\.amazon\.(?:com\.mx|mx|es|co\.uk|com)\/(?:[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+(dp/\w{10})[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+)`,
			allowedDomains: []string{
				// MX site
				"https://www.amazon.com.mx",
				"www.amazon.com.mx",
				"amazon.com.mx",
				//USA site
				"https://www.amazon.com",
				"www.amazon.com",
				"amazon.com",
				// UK site
				"https://www.amazon.co.uk",
				"www.amazon.co.uk",
				"amazon.co.uk",
				// ES site
				"https://www.amazon.es",
				"www.amazon.es",
				"amazon.es",
			},
		},
	}
	for _, f := range options {
		f(&a.shop)
	}

	return a
}

func (a *amazon) ExtractLinks(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	const querySelector = "div.s-main-slot.s-result-list.s-search-results.sg-row a[href]"
	return querySelector, func(e *colly.HTMLElement) {}
}

// GetMetaTags obtiene el contenido de las etiquetas <meta> de la página web
func (a *amazon) GetMetaTags(e *colly.HTMLElement) {}

// GetProductDetails obtiene los datos del producto de la página
func (a *amazon) GetProductDetails(e *colly.HTMLElement) {
	title := e.DOM.Find("span.product-title-word-break").Text()
	title = strings.Trim(title, "\n")
	fmt.Printf("Titulo: %q\n", title)

	brand := e.DOM.Find("div#bylineInfo_feature_div a").Text()
	brand = strings.Trim(brand, "\n")
	fmt.Printf("Marca: %q\n", brand)

	strPrice := e.DOM.
		Find("div#unifiedPrice_feature_div").
		Find(`span#priceblock_ourprice`).
		Text()
	replacer := strings.NewReplacer("$", "", ",", "")
	strPrice = replacer.Replace(strPrice)
	pricef64, err := strconv.ParseFloat(strPrice, 64)
	if err != nil {
		fmt.Printf("Ocurrio un error al parsear el texto del precio: %v", err)
	}
	price := item.ToCurrency(pricef64)
	fmt.Printf("Precio: %q\n", price)
}
