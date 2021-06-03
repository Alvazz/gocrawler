package scraper

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/internal/logging"
	"github.com/leosykes117/gocrawler/pkg/item"
	"golang.org/x/net/html"
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

// Detalles del producto:
// detailBullets_feature_div

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

func (a *amazon) HTMLEvents(evts ...string) []OnHTMLEvent {
	events := make([]OnHTMLEvent, 0)
	for _, funcName := range evts {
		var e OnHTMLEvent
		switch funcName {
		case "GetMetaTags":
			e = a.GetMetaTags
		case "ExtractLinks":
			e = a.ExtractLinks
		case "GetProductDetails":
			e = a.GetProductDetails
		case "GetProductInformation":
			e = a.GetProductInformation
		case "GetProductReviews":
			fmt.Println("Aplicando el evento GetProductReviews")
			e = a.GetProductReviews
		default:
			continue
		}
		events = append(events, e)
	}
	return events
}

func (a *amazon) ExtractLinks(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	const querySelector = "div.s-main-slot.s-result-list.s-search-results.sg-row a[href]"
	return querySelector, func(e *colly.HTMLElement) {
		if onHTML != nil {
			onHTML(e)
		}
	}
}

// GetMetaTags obtiene el contenido de las etiquetas <meta> de la página web
func (a *amazon) GetMetaTags(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	return "html", func(e *colly.HTMLElement) {
		if onHTML != nil {
			onHTML(e)
		}
	}
}

// GetProductDetails obtiene los datos del producto de la página
func (a *amazon) GetProductDetails(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	return "div#centerCol", func(e *colly.HTMLElement) {
		const (
			sourceStore = "Amazon"
		)

		name := e.DOM.Find("span.product-title-word-break").Text()
		name = strings.Trim(name, "\n")
		fmt.Printf("Nombre del producto: %q\n", name)

		brand := e.DOM.Find("div#bylineInfo_feature_div a").Text()
		brand = strings.Trim(brand, "\n")
		fmt.Printf("Marca: %q\n", brand)

		strStars := strings.Fields(e.DOM.Find("i.a-icon.a-icon-star").Text())[0]
		fmt.Printf("Calificación: %q\n", strStars)
		stars, err := strconv.ParseFloat(strStars, 64)
		if err != nil {
			logging.ErrorLogger.Printf("Ocurrio un error al convertir la calificación del producto: %v", err)
		}

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

		productID, err := item.CreateID(name, sourceStore)
		if err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al crear el id del producto %q: %v", productID, err)
		}
		itm := item.NewItem(
			item.ID(productID),
			item.Name(name),
			item.Brand(brand),
			item.Price(pricef64),
			item.Rating(stars),
			item.SourceStore(sourceStore),
			item.URL(e.Request.AbsoluteURL(e.Request.URL.String())),
		)
		itmJSON, err := itm.MarshalJSON()
		if err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al crear el json del producto %q: %v", productID, err)
		}
		e.Request.Ctx.Put("Product", string(itmJSON))

		if onHTML != nil {
			onHTML(e)
		}
	}
}

// GetProductDetails obtiene los datos del producto de la página
func (a *amazon) GetProductInformation(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	classProductDetails := "div#productDetails_feature_div"
	classDetailBullets := "div#detailBulletsWrapper_feature_div"
	return fmt.Sprintf("%s, %s", classProductDetails, classDetailBullets), func(e *colly.HTMLElement) {
		fmt.Println("On GetProductInformation")
		var (
			details item.ProductDetails = make(item.ProductDetails)
		)
		if e.Attr("id") == "detailBulletsWrapper_feature_div" {
			getDetailsFromBulletsWrapper(e.DOM, details)
		} else {
			getDetails(e.DOM, details)
		}
		fmt.Printf("Información del producto: %+q\n", details)

		productJSON := e.Request.Ctx.Get("Product")
		itm := item.NewItem()
		err := itm.UnMarshalJSON(productJSON)
		if err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al formar el struct del producto %q: %v", itm.GetID(), err)
		}
		itm.SetDetails(details)
		itmJSON, err := itm.MarshalJSON()
		if err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al crear el json del producto %q: %v", itm.GetID(), err)
		}
		e.Request.Ctx.Put("Product", string(itmJSON))
	}
}

func getDetailsFromBulletsWrapper(e *goquery.Selection, details item.ProductDetails) {
	e.Find("div#detailBullets_feature_div ul.a-unordered-list li span").Filter(`span.a-text-bold`).Each(func(index int, element *goquery.Selection) {
		m := regexp.MustCompile(`(?m):|\s{2,}|\n+`)
		keyText := m.ReplaceAllString(element.Text(), "")
		nextText := strings.TrimSpace(element.Next().Text())
		details[keyText] = nextText
		fmt.Printf("%d\t%q:%q\n", index, keyText, nextText)
	})
}

func getDetails(e *goquery.Selection, details item.ProductDetails) {
	e.Find(`table.prodDetTable[id*="productDetails"] tr`).Each(func(tableIdx int, row *goquery.Selection) {
		m := regexp.MustCompile(`(?m):|\s{2,}|\n+`)
		keyText := m.ReplaceAllString(row.ChildrenFiltered("th").Text(), "")
		tdNodes := row.ChildrenFiltered("td").Contents()
		valText := tdNodes.FilterFunction(func(i int, el *goquery.Selection) bool {
			return el.Is("span") || el.Nodes[0].Type == html.TextNode
		}).Text()
		valText = m.ReplaceAllString(valText, "")
		details[keyText] = valText
		fmt.Printf("%d\t%q:%q\n", tableIdx, keyText, valText)
	})
}

// a#customerReviews[href="#"] +
// GetProductReviews obtiene los datos de los comentarios hechos por el producto
func (a *amazon) GetProductReviews(onHTML colly.HTMLCallback) (string, colly.HTMLCallback) {
	return `a#customer-reviews-content[href="#"] ~ div.a-row`, func(e *colly.HTMLElement) {
		fmt.Println("On GetProductReviews")
		var (
			reviews item.Comments = make(item.Comments, 0)
		)
		m := regexp.MustCompile(`(?m)\s{2,}|\n+`)
		e.DOM.Find(`div[data-hook="top-customer-reviews-widget"] div.a-section.celwidget`).Each(func(index int, element *goquery.Selection) {
			author := element.Find("span.a-profile-name").Text()
			strStars := strings.Fields(element.Find("i.review-rating span.a-icon-alt").Text())[0]
			title := element.Find(`a[data-hook="review-title"][class*="review-title"] span`).Text()
			_ = element.Find(`span.review-date[data-hook="review-date"]`).Text()
			content := m.ReplaceAllString(element.Find(`span[data-hook="review-body"] div.reviewText[data-hook="review-collapsed"] span`).Text(), "")
			stars, err := strconv.ParseFloat(strStars, 64)
			if err != nil {
				logging.ErrorLogger.Printf("Ocurrio un error al convertir la calificación del producto: %v", err)
			}

			reviews = append(reviews, &item.Comment{
				Title:   title,
				Content: content,
				Author:  author,
				Stars:   item.Score(stars),
			})
		})
		fmt.Printf("Comentarios del producto: %v\n", reviews)

		productJSON := e.Request.Ctx.Get("Product")
		itm := item.NewItem()
		err := itm.UnMarshalJSON(productJSON)
		if err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al formar el struct del producto %q: %v", itm.GetID(), err)
		}
		itm.SetReviews(reviews)
		itmJSON, err := itm.MarshalJSON()
		if err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al crear el json del producto %q: %v", itm.GetID(), err)
		}
		e.Request.Ctx.Put("Product", string(itmJSON))

		if err := a.cacheService.CreateItem(context.Background(), itm); err != nil {
			logging.ErrorLogger.Fatalf("Ocurrio un error al guardar el producto %s: %v", itm.GetID(), err)
		}
	}
}
