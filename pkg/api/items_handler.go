package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/itemparser"
)

func (s *Services) ParserItems() error {
	var wg sync.WaitGroup
	ex := itemparser.NewExtractor(s.itemCacheService)
	errorsCh := make(chan error)
	itemsCh := make(chan *item.Item)
	items := make(item.Items, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	wg.Add(2)
	go func() {
		for err := range errorsCh {
			fmt.Printf("ERROR: %v", err)
		}
		wg.Done()
	}()

	go func() {
		for i := range itemsCh {
			if i != nil {
				fmt.Println("Eliminando el producto", i.GetID())
				items = append(items, i)
			}
		}
		wg.Done()
	}()

	err := ex.ScanItems(ctx, errorsCh, itemsCh)
	wg.Wait()
	if err != nil {
		return err
	}

	log.Printf("Número de productos obtenidos/eliminados: %d\n", len(items))

	itemsJSON, err := items.MarshalJSON()
	if err != nil {
		fmt.Println("Ocurrio un error al hacer marshal de los items")
	}
	jsonFile := "/tmp/item_from_cache.json"
	err = ioutil.WriteFile(jsonFile, itemsJSON, 0600)
	if err != nil {
		return fmt.Errorf("Ocurrio un error al crear el archivo json: %v", err)
	}

	return nil
}
