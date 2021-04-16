package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/logging"
	_ "github.com/lib/pq"
)

type itemRepository struct {
	pool *redis.Pool
}

func NewRepository(pool *redis.Pool) *itemRepository {
	return &itemRepository{
		pool: pool,
	}
}

func (r *itemRepository) CreateItem(ctx context.Context, item *item.Item) error {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}

	productKey := fmt.Sprintf("product:%s", item.ID)
	commentsKey := fmt.Sprintf("comments:%s", item.ID)
	detailsKey := fmt.Sprintf("details:%s", item.ID)

	// SE CREA LA TRANSACCION
	err = conn.Send("MULTI")
	if err != nil {
		return err
	}
	// SE CREA EL HASH PRINCIPAL PARA ALMACENAR EL PRODUCTO
	err = conn.Send("HMSET", productKey, "id", item.ID, "name", item.Name, "brand", item.Brand, "description", item.Description, "score", item.Rating, "reviews", commentsKey, "sourceStore", item.SourceStore, "details", detailsKey)
	if err != nil {
		return err
	}

	// CREAR UNA LISTA DONDE LOS VALORES APUNTAN A UN HASH QUE CONTIENE
	// LOS DATOS DEL COMENTARIO
	for i, comment := range item.Reviews {
		commentKey := fmt.Sprintf("comment:%d:%s", i, item.ID)
		err = conn.Send("RPUSH", commentsKey, commentKey)
		if err != nil {
			return err
		}
		err = conn.Send("HMSET", commentKey, "tile", comment.Title, "content", comment.Content, "author", comment.Author, "stars", comment.Stars, "date", comment.Date.Format("02/01/2006 15:04:05"))
		if err != nil {
			return err
		}
	}

	// ALMACENA EL MAP DE LOS DETALLES DEL PRODUCTO
	for k, v := range item.Description {
		err = conn.Send("HSETNX", detailsKey, k, v)
		if err != nil {
			return err
		}
	}

	// EJECUTA LA TRANSACCION
	results, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return err
	}

	errs := make([]string, 0)

	for i, result := range results {
		v, ok := result.(error)
		if ok {
			errs = append(errs, fmt.Sprintf("%v", v))
		} else {
			logging.InfoLogger.Printf("[%s]Respuesta del comando %d: %v", item.ID, i, v)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (r *itemRepository) FetchItemID(ctx context.Context, ID string) (*item.Item, error) {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	result, err := redis.String(conn.Do("GET", ID))
	if err != nil {
		return nil, err
	}

	if result == "" {
		return nil, errors.New("not found")
	}

	gopher := &item.Item{}
	err = json.Unmarshal([]byte(result), gopher)

	return gopher, err
}
