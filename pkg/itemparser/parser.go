package itemparser

import (
	"context"
	"fmt"

	"github.com/leosykes117/gocrawler/pkg/item"
)

type itemParser struct {
	item         *item.Item
	cacheService *item.CacheService
	ctx          context.Context
	errorsCh     chan<- error
	itemsCh      chan<- *item.Item
}

func NewItemParser(ctx context.Context, id string, cacheService *item.CacheService, errorsCh chan<- error, itemsCh chan<- *item.Item) *itemParser {
	return &itemParser{
		item:         &item.Item{ID: id},
		cacheService: cacheService,
		ctx:          ctx,
		errorsCh:     errorsCh,
		itemsCh:      itemsCh,
	}
}

func (parser *itemParser) Task() {
	var err error
	parser.item, err = parser.cacheService.FetchItemID(parser.ctx, parser.item.ID)
	if err != nil {
		parser.errorsCh <- err
		return
	}

	fmt.Printf("Analizando los comentarios de %q\n", parser.item.ID)
	if len(parser.item.Reviews) > 0 {
		reviewsSentiment := anlz.AnalyzeComments(parser.item.ID, parser.item.Reviews)
		if len(reviewsSentiment) > 0 {
			fmt.Printf("reviews of %q: %v\n", parser.item.ID, reviewsSentiment)

		}
	}

	parser.itemsCh <- parser.item
}
