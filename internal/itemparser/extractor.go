package itemparser

import (
	"context"
	"fmt"

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
	err    error
	keys   []string
)

// GetItemsFromCache obtiene n cantidad de items almacenado en Redis
func GetItemsFromCache() (map[string]interface{}, error) {
	repo := redis.NewRepository(redis.NewConn(":6379"))
	for {
		ctx := context.Background()
		items, cursor, err = repo.FetchTopItems(ctx, cursor, scanCount)
		if err != nil {
			return nil, fmt.Errorf("Error al obtener los productos de redis: %v", err)
		}

		//file, err := os.OpenFile(filepath.Join("/Users/leonardo/", "crawling-data", "outs", "products.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		//defer file.Close()
		if err != nil {
			return nil, fmt.Errorf("Error al crear el archivo: %v", err)
		}

		//datawriter := bufio.NewWriter(file)
		//defer datawriter.Flush()

		/* if _, err := datawriter.WriteString(fmt.Sprintf("ITERACION %d\n", loop)); err != nil {
			fmt.Printf("Error al escribir en el archivo: %v", err)
		} */

		if len(items) > 0 {
			count += len(items)
			keys = make([]string, 0)
			for _, item := range items {
				fmt.Printf("item => %+v\n", item)
				productKey := fmt.Sprintf("product:%s", item.ID)
				commentsKey := fmt.Sprintf("comments:%s", item.ID)
				detailsKey := fmt.Sprintf("details:%s", item.ID)
				keys = append(keys, productKey, commentsKey, detailsKey)
				//datawriter.WriteString(fmt.Sprintf("%s\n%s\n%s\n\n", productKey, commentsKey, detailsKey))
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
		//datawriter.WriteString("\n\n\n")
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
