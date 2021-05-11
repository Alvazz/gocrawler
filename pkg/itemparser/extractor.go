package itemparser

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hako/durafmt"
	"github.com/leosykes117/gocrawler/internal/env"
	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/storage/redis"
)

const (
	scanCount = 200
)

// GetItemsFromCache obtiene n cantidad de items almacenado en Redis
func GetItemsFromCache() (map[string]interface{}, error) {
	var (
		items  item.Items
		cursor int = 0
		count  int = 0
		loop   int = 0
		keys   []string
	)

	if _, ok := os.LookupEnv("GO_CRAWLER_REDIS_ENDPOINT"); !ok {
		fmt.Println("Leyendo las variables del archivo")
		if err := env.LoadVars(); err != nil {
			return nil, fmt.Errorf("Error al establecer las variables: %v", err)
		}
	}

	err := env.ReadVars()
	if err != nil {
		return nil, fmt.Errorf("Error al leer la configuración: %v", err)
	}

	endpoint, err := env.GetCrawlerVars(env.RedisEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener el endpoint de redis: %v", err)
	}

	repo := redis.NewRepository(redis.NewConn(endpoint.(string)))
	anlz := NewAnalyzer()
	for {
		ctx := context.Background()
		startFetch := time.Now()
		items, cursor, err = repo.FetchTopItems(ctx, cursor, scanCount)
		elapsed := time.Since(startFetch)
		fmt.Println("Tiempo FetchTopItems:", durafmt.Parse(elapsed))
		if err != nil {
			return nil, fmt.Errorf("Error al obtener los productos de redis: %v", err)
		}
		fmt.Println("Número de productos obtenidos:", len(items))
		if len(items) > 0 {
			count += len(items)
			keys = make([]string, 0)
			for _, itm := range items {
				//fmt.Printf("item => %+v\n", itm)
				productKey := fmt.Sprintf("product:%s", itm.ID)
				commentsKey := fmt.Sprintf("comments:%s", itm.ID)
				detailsKey := fmt.Sprintf("details:%s", itm.ID)
				keys = append(keys, productKey, commentsKey, detailsKey)

				for i := 0; i < len(itm.Reviews); i++ {
					commentKey := fmt.Sprintf("comment:%d:%s", i, itm.ID)
					keys = append(keys, commentKey)
				}

				if loop == 0 {
					reviewAnalysis := anlz.AnalyzeComments(itm.ID, itm.Reviews)
					fmt.Println("Numero de comentarios analizados:", len(reviewAnalysis))
					for k, v := range reviewAnalysis {
						fmt.Printf("%s, %s\n", k, v.GoString())
					}
				}
			}
			err := repo.Delete(ctx, keys...)
			fmt.Println("Productos eliminados:", len(keys))
			if err != nil {
				return nil, fmt.Errorf("Error al eliminar los productos de redis: %v", err)
			}
		}

		loop++

		if cursor == 0 {
			fmt.Println("La iteración terminó")
			fmt.Println("Items:", count)
			fmt.Println("Iteraciones:", loop)
			break
		}
	}

	return map[string]interface{}{
		"count": count,
		"loops": loop,
	}, nil
}
