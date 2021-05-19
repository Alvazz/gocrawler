package api

import (
	"fmt"
	"os"

	"github.com/leosykes117/gocrawler/internal/env"
	"github.com/leosykes117/gocrawler/pkg/item"
	"github.com/leosykes117/gocrawler/pkg/storage"
	"github.com/leosykes117/gocrawler/pkg/storage/redis"
)

type Services struct {
	itemCacheService *item.CacheService
}

func NewServices() (*Services, error) {
	err := config()
	if err != nil {
		return nil, err
	}
	storage.New(storage.Redis)
	services := &Services{
		itemCacheService: item.NewCacheService(redis.NewRepository(storage.MemoryPool())),
	}
	return services, nil
}

func config() error {
	if _, ok := os.LookupEnv("GO_CRAWLER_REDIS_ENDPOINT"); !ok {
		fmt.Println("Leyendo las variables del archivo")
		if err := env.LoadVars(); err != nil {
			return fmt.Errorf("Error al establecer las variables: %v", err)
		}
	}

	err := env.ReadVars()
	if err != nil {
		return fmt.Errorf("Error al leer la configuración: %v", err)
	}
	return nil
}
