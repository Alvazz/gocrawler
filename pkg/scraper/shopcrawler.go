package scraper

import (
	"github.com/leosykes117/gocrawler/pkg/item"
)

type shop struct {
	cacheService        *item.CacheService
	topLevelDomain      string
	keywordsValue       string
	descriptionValue    string
	linkExtractionQuery string
	linkProductQuery    string
	domainGlob          string
	allowedDomains      []string
}

func cacheService(cs *item.CacheService) func(*shop) {
	return func(s *shop) {
		s.cacheService = cs
	}
}

func topLevelDomain(tld string) func(*shop) {
	return func(s *shop) {
		s.topLevelDomain = tld
	}
}

func descriptionValue(dv string) func(*shop) {
	return func(s *shop) {
		s.descriptionValue = dv
	}
}

func linkExtractionQuery(leq string) func(*shop) {
	return func(s *shop) {
		s.linkExtractionQuery = leq
	}
}

func linkProductQuery(lpq string) func(*shop) {
	return func(s *shop) {
		s.linkProductQuery = lpq
	}
}

func domainGlob(dg string) func(*shop) {
	return func(s *shop) {
		s.domainGlob = dg
	}
}

func allowedDomains(ad ...string) func(*shop) {
	return func(s *shop) {
		s.allowedDomains = ad
	}
}

func (sp *shop) GetLinkExtractionQuery() string {
	return sp.linkExtractionQuery
}

func (sp *shop) GetLinkProductQuery() string {
	return sp.linkProductQuery
}

func (sp *shop) GetAllowedDomains() []string {
	return sp.allowedDomains
}

func (sp *shop) GetDomainGlob() string {
	return sp.domainGlob
}
