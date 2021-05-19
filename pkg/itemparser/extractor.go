package itemparser

import (
	"context"
	"fmt"

	"github.com/leosykes117/gocrawler/pkg/item"
)

const (
	scanCount = 200
)

type extractor struct {
	cacheService *item.CacheService
	itemsCount   int
	scanLoops    int
}

func NewExtractor(cacheService *item.CacheService) *extractor {
	return &extractor{
		cacheService: cacheService,
		itemsCount:   0,
		scanLoops:    0,
	}
}

// scanItems realiza un ite
func (e *extractor) ScanItems(ctx context.Context, errorsCh chan<- error, itemsCh chan<- *item.Item) error {
	var (
		cursor    int = 0
		itemsKeys []string
		err       error
	)
	NewAnalyzer()
	for {
		// TODO 💣: pensar bien lo del ctx
		itemsKeys, cursor, err = e.cacheService.ScanItems(ctx, cursor, scanCount)
		if err != nil {
			return err
		}

		fmt.Println("Número de productos obtenidos:", len(itemsKeys))
		e.itemsCount += len(itemsKeys)
		e.scanLoops++

		w := NewWorkPool(len(itemsKeys))

		for _, k := range itemsKeys {
			w.Run(NewItemParser(ctx, k, e.cacheService, errorsCh, itemsCh))
		}

		w.Shutdown()

		if cursor == 0 {
			break
		}
	}
	fmt.Println("La iteración terminó")
	fmt.Println("Items:", e.itemsCount)
	fmt.Println("Iteraciones:", e.scanLoops)
	close(itemsCh)
	close(errorsCh)
	return nil
}
