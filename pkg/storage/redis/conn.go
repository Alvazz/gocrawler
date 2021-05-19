package redis

import (
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/leosykes117/gocrawler/internal/env"
)

var once sync.Once

// NewConn establece la conexión con redis retornando un pool de conexión.
func NewConn() (pool *redis.Pool) {
	once.Do(func() {
		endpoint, _ := env.GetCrawlerVars(env.RedisEndpoint)
		port, _ := env.GetCrawlerVars(env.RedisPort)
		fmt.Printf("Endpoint:%q\nPort:%q\n", endpoint, port)
		addr := fmt.Sprintf("%s:%s", endpoint, port)
		pool = &redis.Pool{
			MaxIdle:     50,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", addr, redis.DialConnectTimeout(time.Minute))
				if err != nil {
					fmt.Printf("REDIS CONNECTION ERROR: %v", err)
					return nil, err
				}
				fmt.Println("REDIS: Conexión establecida")
				return c, nil
			},
		}
	})
	return
}
