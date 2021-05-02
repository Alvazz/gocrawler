package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// NewConn establece la conexión con redis retornando un pool de conexión.
func NewConn(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr, redis.DialConnectTimeout(time.Minute))
			if err != nil {
				fmt.Printf("REDIS CONNECTION ERROR: %v", err)
				return nil, err
			}
			fmt.Println("Conectado a redis")
			return c, nil
		},
	}
}
