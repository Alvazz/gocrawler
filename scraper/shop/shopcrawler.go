package shop

import (
	"github.com/gocolly/colly"
)

type ShopCrawler interface {
	GetMetaTags(*colly.HTMLElement)
	GetProductData(*colly.HTMLElement)
	GetProductDetails(*colly.HTMLElement)
}

type shop struct {
	topLevelDomain      string
	keywordsValue       string
	descriptionValue    string
	linkExtractionQuery string
}

func (sp *shop) GetLinkExtractionQuery() string {
	return sp.linkExtractionQuery
}
