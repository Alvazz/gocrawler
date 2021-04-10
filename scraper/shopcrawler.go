package scraper

import (
	"github.com/gocolly/colly"
)

type shopCrawler interface {
	GetMetaTags(*colly.HTMLElement)
	GetProductDetails(*colly.HTMLElement, *Scraper)
	GetLinkExtractionQuery() string
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
