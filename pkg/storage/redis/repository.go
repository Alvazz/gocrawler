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

// NewRepository retorna un objeto que implementa la interfaz item.Cache
func NewRepository(pool *redis.Pool) item.Cache {
	return &itemRepository{pool}
}

func (r *itemRepository) Set(ctx context.Context, item *item.Item) error {
	var err error
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	productID := item.GetID()
	productKey := fmt.Sprintf("product:%s", productID)
	commentsKey := fmt.Sprintf("comments:%s", productID)
	detailsKey := fmt.Sprintf("details:%s", productID)
	imagesKey := fmt.Sprintf("images:%s", productID)

	// SE CREA LA TRANSACCION
	err = conn.Send("MULTI")
	if err != nil {
		return err
	}
	// SE CREA EL HASH PRINCIPAL PARA ALMACENAR EL PRODUCTO
	err = conn.Send("HMSET", productKey, "id", productID, "name", item.GetName(), "brand", item.GetBrand(), "description", item.GetDescription(), "price", item.GetPrice(), "score", item.GetRating(), "reviews", commentsKey, "sourceStore", item.GetSourceStore(), "url", item.GetURL(), "details", detailsKey, "images", imagesKey)
	if err != nil {
		return err
	}

	// CREAR UNA LISTA DONDE LOS VALORES APUNTAN A UN HASH QUE CONTIENE
	// LOS DATOS DEL COMENTARIO
	for i, comment := range item.GetReviews() {
		commentKey := fmt.Sprintf("comment:%d:%s", i, productID)
		err = conn.Send("RPUSH", commentsKey, commentKey)
		if err != nil {
			return err
		}
		err = conn.Send("HMSET", commentKey, "title", comment.Title, "content", comment.Content, "author", comment.Author, "stars", comment.Stars, "date", comment.Date.Format("02/01/2006 15:04:05"))
		if err != nil {
			return err
		}
	}

	// ALMACENA EL MAP DE LOS DETALLES DEL PRODUCTO
	for k, v := range item.GetDetails() {
		err = conn.Send("HSETNX", detailsKey, k, v)
		if err != nil {
			return err
		}
	}

	// CREAR UNA LISTA DE LA IMAGENES
	for _, image := range item.GetImages() {
		err = conn.Send("RPUSH", imagesKey, image)
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
			fmt.Printf("[%s]Respuesta del comando %d: %v\n", item.GetID(), i, v)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func (r *itemRepository) Get(ctx context.Context, ID string) (*item.Item, error) {
	var (
		id, name, brand, description, sourceStore, url string
		rating, price                                  float64
		reviews                                        item.Comments
		details                                        item.ProductDetails
		images                                         []string
		err                                            error
	)

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener la conexión: %v", err)
	}

	result, err := redis.StringMap(conn.Do("HGETALL", ID))
	conn.Close()
	if err != nil {
		return nil, fmt.Errorf("Error en el HGETALL: %v", err)
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
		case "price":
			price, _ = strconv.ParseFloat(v, 64)
		case "sourceStore":
			sourceStore = v
		case "url":
			url = v
		case "score":
			rating, _ = strconv.ParseFloat(v, 64)
		case "reviews":
			reviews, err = r.FetchReviews(ctx, nil, v)
			if err != nil {
				return nil, fmt.Errorf("Error al obtener los comentarios: %v", err)
			}
		case "details":
			details, err = r.FetchItemDetails(ctx, nil, v)
			if err != nil {
				return nil, fmt.Errorf("Error al obtener los detalles del producto: %v", err)
			}
		case "images":
			images, err = r.FetchItemImages(ctx, nil, v)
			if err != nil {
				return nil, fmt.Errorf("Error al obtener los detalles del producto: %v", err)
			}
		}
	}

	i := item.NewItem(
		item.ID(id),
		item.Name(name),
		item.Brand(brand),
		item.Description(description),
		item.Price(price),
		item.SourceStore(sourceStore),
		item.URL(url),
		item.Rating(rating),
		item.Reviews(reviews),
		item.Details(details),
		item.Images(images),
	)
	return i, nil
}

func (r *itemRepository) FetchItemDetails(ctx context.Context, conn redis.Conn, detailID string) (item.ProductDetails, error) {
	var err error
	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
		defer conn.Close()
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
		reviews item.Comments
		err     error
	)
	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
		defer conn.Close()
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
		comment, err := r.GetCommentByID(ctx, conn, k)
		if err != nil {
			fmt.Printf("Ocurrio un error al obtener el comentario %s: %v\n", commentID, err)
		}
		reviews = append(reviews, comment)
	}
	return reviews, nil
}

func (r *itemRepository) GetCommentByID(ctx context.Context, conn redis.Conn, commentID string) (*item.Comment, error) {
	var (
		title, content, author string
		stars                  float64
		date                   time.Time
		err                    error
	)

	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
		defer conn.Close()
	}
	if err != nil {
		return nil, err
	}

	commentData, err := redis.StringMap(conn.Do("HGETALL", commentID))
	if err != nil {
		return nil, err
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
	return item.NewComment(title, content, author, item.Score(stars), date), nil
}

func (r *itemRepository) Scan(ctx context.Context, cursor int, n int) ([]string, int, error) {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, cursor, err
	}
	defer conn.Close()

	arr, err := redis.Values(conn.Do("SCAN", cursor, "MATCH", "product:*", "COUNT", n))
	if err != nil {
		return nil, cursor, err
	}

	if len(arr) == 0 {
		return nil, cursor, nil
	}

	cursor, _ = redis.Int(arr[0], nil)
	keys, _ := redis.Strings(arr[1], nil)
	fmt.Println("Cursor:", cursor)
	return keys, cursor, nil
}

func (r *itemRepository) Delete(ctx context.Context, keys ...string) error {
	conn, err := r.pool.GetContext(ctx)
	defer conn.Close()
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

func (r *itemRepository) FetchItemImages(ctx context.Context, conn redis.Conn, imagesKey string) ([]string, error) {
	var (
		err error
	)
	if conn == nil {
		conn, err = r.pool.GetContext(ctx)
		defer conn.Close()
	}
	if err != nil {
		return nil, err
	}
	images, err := redis.Strings(conn.Do("LRANGE", imagesKey, 0, -1))
	if err != nil {
		return nil, err
	}

	return images, nil
}
