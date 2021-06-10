package env

import "fmt"

type GoCrawlerSpecification struct {
	RedisEndpoint string `envconfig:"GO_CRAWLER_REDIS_ENDPOINT"`
	RedisPort     string `envconfig:"GO_CRAWLER_REDIS_PORT"`
	DebugRequests bool   `envconfig:"GO_CRAWLER_DEBUG_REQUESTS"`
	SeedURL       string `envconfig:"GO_CRAWLER_SEEDURL"`
}

const (
	RedisEndpoint = "REDIS_ENDPOINT"
	RedisPort     = "REDIS_PORT"
	DebugRequests = "DEBUG_REQUESTS"
	SeedURL       = "SEEDURL"
)

var (
	crawlerEnvVars GoCrawlerSpecification
)

// GetCrawlerVars devuelve el valor de la variable de ambiente especificada en envar
func GetCrawlerVars(envar string) (interface{}, error) {
	switch envar {
	case RedisEndpoint:
		return crawlerEnvVars.RedisEndpoint, nil
	case RedisPort:
		return crawlerEnvVars.RedisPort, nil
	case DebugRequests:
		return crawlerEnvVars.DebugRequests, nil
	case SeedURL:
		return crawlerEnvVars.SeedURL, nil
	default:
		return "", fmt.Errorf("No existe la variable %s", envar)
	}
}

func SetCrawlerVars(envar string, val interface{}) error {
	switch envar {
	case SeedURL:
		crawlerEnvVars.SeedURL = val.(string)
	default:
		return fmt.Errorf("No se puede modificar el valor de %s", envar)
	}
	return nil
}
