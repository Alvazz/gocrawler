package itemparser

import (
	"context"
	"fmt"

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

	err := env.ReadVars()
	if err != nil {
		return nil, fmt.Errorf("Error al leer la configuración: %v", err)
	}

	endpoint, err := env.GetEnvs(env.RedisEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error al obtenerl el endpoint de redis: %v", err)
	}

	repo := redis.NewRepository(redis.NewConn(endpoint.(string)))
	anlz := NewAnalyzer()
	for {
		ctx := context.Background()
		items, cursor, err = repo.FetchTopItems(ctx, cursor, scanCount)
		if err != nil {
			return nil, fmt.Errorf("Error al obtener los productos de redis: %v", err)
		}

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
				reviewAnalysis := anlz.AnalyzeComments(itm.ID, itm.Reviews)
				for k, v := range reviewAnalysis {
					fmt.Printf("%s, %s", k, v.GoString())
				}
			}
			/* err := repo.Delete(ctx, keys...)
			fmt.Println("Productos eliminados:", len(keys))
			if err != nil {
				return nil, fmt.Errorf("Error al eliminar los productos de redis: %v", err)
			} */
		}

		loop++

		if cursor == 0 {
			fmt.Println("La iteración terminó")
			fmt.Println("Items:", count)
			break
		}
	}

	return map[string]interface{}{
		"count": count,
		"loops": loop,
	}, nil
}
