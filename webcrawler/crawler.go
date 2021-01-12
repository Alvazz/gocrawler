package webcrawler

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func Start() {
	c := colly.NewCollector(
		colly.AllowedDomains("amazon.com.mx"),
		colly.Async(true),
	)

	// Callback for links on scraped pages
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the linked URL from the anchor tag
		link := e.Attr("href")
		// Have our crawler visit the linked URL
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// 
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})
}
