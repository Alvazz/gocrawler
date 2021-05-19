package item

import "context"

type Cache interface {
	// CreateItem perciste un producto
	Set(context.Context, *Item) error
	// FetchItemID obtiene un producto por su ID
	Get(context.Context, string) (*Item, error)
	// ScanItems obtiene un producto por su ID
	Scan(context.Context, int, int) ([]string, int, error)
	//Flush elimina todas las llaves relacionadas al producto
	Delete(context.Context, ...string) error
}

type CacheService struct {
	memoryDB Cache
}

func NewCacheService(cache Cache) *CacheService {
	return &CacheService{cache}
}

func (c *CacheService) CreateItem(ctx context.Context, item *Item) error {
	return c.memoryDB.Set(ctx, item)
}

func (c *CacheService) FetchItemID(ctx context.Context, ID string) (*Item, error) {
	return c.memoryDB.Get(ctx, ID)
}

func (c *CacheService) ScanItems(ctx context.Context, cursor int, count int) ([]string, int, error) {
	return c.memoryDB.Scan(ctx, cursor, count)
}

func (c *CacheService) Delete(ctx context.Context, keys ...string) error {
	return c.memoryDB.Delete(ctx, keys...)
}
