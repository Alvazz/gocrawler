package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/leosykes117/gocrawler/pkg/item"
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
	var err error
	conn := r.pool.Get()
	defer r.pool.Close()

	productKey := fmt.Sprintf("product:%s", item.ID)
	commentsKey := fmt.Sprintf("comments:%s", item.ID)
	detailsKey := fmt.Sprintf("details:%s", item.ID)

	// SE CREA LA TRANSACCION
	err = conn.Send("MULTI")
	if err != nil {
		return err
	}
	// SE CREA EL HASH PRINCIPAL PARA ALMACENAR EL PRODUCTO
	err = conn.Send("HMSET", productKey, "id", item.ID, "name", item.Name, "brand", item.Brand, "description", item.Description, "score", item.Rating, "reviews", commentsKey, "sourceStore", item.SourceStore, "url", item.URL, "details", detailsKey)
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
	for k, v := range item.Details {
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
			fmt.Printf("[%s]Respuesta del comando %d: %v", item.ID, i, v)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (r *itemRepository) FetchItemID(ctx context.Context, conn redis.Conn, ID string) (*item.Item, error) {
	var (
		id, name, brand, description, sourceStore, url string
		rating                                         float64
		reviews                                        item.Comments
		details                                        item.ProductDetails
		err                                            error
	)

	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
	}

	if err != nil {
		return nil, err
	}

	result, err := redis.StringMap(conn.Do("HGETALL", ID))
	if err != nil {
		return nil, err
	}

	for k, v := range result {
		switch k {
		case "id":
			id = v
		case "name":
			name = v
		case "brand":
			brand = v
		case "description":
			description = v
		case "sourceStore":
			sourceStore = v
		case "url":
			url = v
		case "score":
			rating, _ = strconv.ParseFloat(v, 64)
		case "reviews":
			reviews, err = r.FetchReviews(ctx, conn, v)
			if err != nil {
				return nil, err
			}
		case "details":
			details, err = r.FetchItemDetails(ctx, conn, v)
			if err != nil {
				return nil, err
			}
		}
	}

	i := item.NewItem(name, brand, description, sourceStore, url, item.Score(rating), reviews, details)
	i.ID = id
	return i, nil
}

func (r *itemRepository) FetchItemDetails(ctx context.Context, conn redis.Conn, detailID string) (item.ProductDetails, error) {
	var err error
	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
	}

	if err != nil {
		return nil, err
	}

	itemDetails, err := redis.StringMap(conn.Do("HGETALL", detailID))
	if err != nil {
		return nil, err
	}

	return itemDetails, nil
}

func (r *itemRepository) FetchReviews(ctx context.Context, conn redis.Conn, commentID string) (item.Comments, error) {
	var (
		title, content, author string
		stars                  float64
		date                   time.Time
		reviews                item.Comments
		err                    error
	)
	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
	}
	if err != nil {
		return nil, err
	}

	result, err := redis.Strings(conn.Do("LRANGE", commentID, 0, -1))
	if err != nil {
		return nil, err
	}

	reviews = make(item.Comments, 0, len(result))

	for _, k := range result {
		commentData, err := redis.StringMap(conn.Do("HGETALL", k))
		if err != nil {
			fmt.Printf("Ocurrio un error al obtener el comentario %s\n", commentID)
		}
		for key, val := range commentData {
			switch key {
			case "tile":
				title = val
			case "content":
				content = val
			case "author":
				author = val
			case "stars":
				stars, _ = strconv.ParseFloat(val, 64)
			case "date":
				date, _ = time.Parse("02/01/2006 15:04:05", val)
			}
		}
		reviews = append(reviews, item.NewComment(title, content, author, item.Score(stars), date))
	}
	return reviews, nil
}

func (r *itemRepository) FetchTopItems(ctx context.Context, cursor int, n int) (item.Items, int, error) {
	var items item.Items
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, cursor, err
	}

	arr, err := redis.Values(conn.Do("SCAN", strconv.Itoa(cursor), "MATCH", "product:*", "COUNT", strconv.Itoa(n)))
	if err != nil {
		return nil, cursor, err
	}

	if len(arr) == 0 {
		return item.Items{}, cursor, nil
	}

	cursor, _ = redis.Int(arr[0], nil)
	fmt.Println("\n\nCursor:", cursor)

	keys, _ := redis.Strings(arr[1], nil)
	for _, k := range keys {
		i, err := r.FetchItemID(ctx, conn, k)
		if err != nil {
			fmt.Printf("ERROR %s => %v", k, err)
		}
		items = append(items, i)
	}

	return items, cursor, nil
}

func (r *itemRepository) Delete(ctx context.Context, keys ...string) error {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		args = append(args, key)
	}

	_, err = conn.Do("DEL", args...)
	return err
}
