package scraper

import (
	"github.com/gocolly/colly"
	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/storage"
	"github.com/leosykes117/gocrawler/pkg/storage/redis"
)

type shopCrawler interface {
	GetMetaTags(*colly.HTMLElement)
	GetProductDetails(*colly.HTMLElement)
	GetLinkExtractionQuery() string
	GetLinkProductQuery() string
	GetAllowedDomains() []string
	GetDomainGlob() string
}

const (
	Mixup  = "MIXUP"
	Amazon = "AMAZON"
)

func ShopFactory(store string) shopCrawler {
	storage.New(storage.Redis)
	s := item.NewCacheService(redis.NewRepository(storage.MemoryPool()))
	switch store {
	case Mixup:
		return newShopMixup(cacheService(s))
	case Amazon:
		return nil
	default:
		return nil
	}
}
