package storage

import (
	"log"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/leosykes117/gocrawler/pkg/storage/redis"
)

var (
	memoryDB *redigo.Pool
)

// Driver es de tipo string e indica el el motor de base de datos a utilizar
type Driver string

const (
	// Redis base de datos en memoria.
	Redis Driver = "REDIS"
	// PostgreSQL base de datos para el catalogo de productos.
	PostgreSQL Driver = "POSTGRES"
)

// New crea la conexión con la base de datos
func New(d Driver) {
	switch d {
	case Redis:
		memoryDB = redis.NewConn()
	default:
		log.Fatalf("El driver %s no esta implementado", d)
	}
}

// MemoryPool retorna una unica instancia de memoryDB
func MemoryPool() *redigo.Pool {
	return memoryDB
}

// TODO: 💣 Implementar el facory retornando la interfaz correspondiente
