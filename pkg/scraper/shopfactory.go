package scraper

import (
	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/storage"
	"github.com/leosykes117/gocrawler/pkg/storage/redis"
)

type OnHTMLEvent func(colly.HTMLCallback) (string, colly.HTMLCallback)

type shopCrawler interface {
	HTMLEvents(...string) []OnHTMLEvent
	GetMetaTags(colly.HTMLCallback) (string, colly.HTMLCallback)
	GetProductDetails(colly.HTMLCallback) (string, colly.HTMLCallback)
	ExtractLinks(colly.HTMLCallback) (string, colly.HTMLCallback)
	GetLinkExtractionQuery() string
	GetLinkProductQuery() string
	GetAllowedDomains() []string
	GetDomainGlob() string
}

const (
	Mixup  = "MIXUP"
	Amazon = "AMAZON"
)

// shopFactory es el factory de shopCrawler
func shopFactory(store string) shopCrawler {
	storage.New(storage.Redis)
	s := item.NewCacheService(redis.NewRepository(storage.MemoryPool()))
	switch store {
	case Mixup:
		return newShopMixup(cacheService(s))
	case Amazon:
		return newShopAmazon(cacheService(s))
	default:
		return nil
	}
}
