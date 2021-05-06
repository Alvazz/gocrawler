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

var (
	items  item.Items
	cursor int = 0
	count  int = 0
	loop   int = 0
	keys   []string
)

// GetItemsFromCache obtiene n cantidad de items almacenado en Redis
func GetItemsFromCache() (map[string]interface{}, error) {
	err := env.ReadVars()
	if err != nil {
		return nil, fmt.Errorf("Error al leer la configuración: %v", err)
	}

	endpoint, err := env.GetEnvs(env.RedisEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error al obtenerl el endpoint de redis: %v", err)
	}

	repo := redis.NewRepository(redis.NewConn(endpoint.(string)))
	for {
		ctx := context.Background()
		items, cursor, err = repo.FetchTopItems(ctx, cursor, scanCount)
		if err != nil {
			return nil, fmt.Errorf("Error al obtener los productos de redis: %v", err)
		}

		if err != nil {
			return nil, fmt.Errorf("Error al crear el archivo: %v", err)
		}

		if len(items) > 0 {
			count += len(items)
			keys = make([]string, 0)
			for _, item := range items {
				fmt.Printf("item => %+v\n", item)
				productKey := fmt.Sprintf("product:%s", item.ID)
				commentsKey := fmt.Sprintf("comments:%s", item.ID)
				detailsKey := fmt.Sprintf("details:%s", item.ID)
				keys = append(keys, productKey, commentsKey, detailsKey)

				for i := 0; i < len(item.Reviews); i++ {
					commentKey := fmt.Sprintf("comment:%d:%s", i, item.ID)
					keys = append(keys, commentKey)
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
			break
		}
	}

	return map[string]interface{}{
		"count": count,
		"loops": loop,
	}, nil
}
