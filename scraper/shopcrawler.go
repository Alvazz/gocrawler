package scraper

import (
	"github.com/gocolly/colly"
)

type shopCrawler interface {
	GetMetaTags(*colly.HTMLElement)
	GetProductDetails(*colly.HTMLElement, *Scraper)
	GetLinkExtractionQuery() string
	GetLinkProductQuery() string
}

type shop struct {
	topLevelDomain      string
	keywordsValue       string
	descriptionValue    string
	linkExtractionQuery string
	linkProductQuery    string
}

func (sp *shop) GetLinkExtractionQuery() string {
	return sp.linkExtractionQuery
}

func (sp *shop) GetLinkProductQuery() string {
	return sp.linkProductQuery
}
