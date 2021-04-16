package redis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gomodule/redigo/redis"
	"github.com/leosykes117/gocrawler/scraper"
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

func (r *itemRepository) CreateItem(ctx context.Context, item *scraper.Item) error {
	bytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", item.ID, string(bytes))
	return err
}

func (r *itemRepository) FetchItemID(ctx context.Context, ID string) (*scraper.Item, error) {
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

	gopher := &scraper.Item{}
	err = json.Unmarshal([]byte(result), gopher)

	return gopher, err
}
