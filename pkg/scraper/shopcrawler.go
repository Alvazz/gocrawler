package scraper

import (
	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/pkg/item"
)

type shopCrawler interface {
	GetMetaTags(*colly.HTMLElement)
	GetProductDetails(*colly.HTMLElement, *Scraper)
	GetLinkExtractionQuery() string
	GetLinkProductQuery() string
}

type shop struct {
	chacheService       *item.CacheService
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
