package env

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/leosykes117/gocrawler/internal/logging"
)

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
	envFilePath    string
)

// LoadVars carga el archivo .env en variables de ambiente
func LoadVars() error {
	var err error
	projectPath, ok := os.LookupEnv("PROJECTPATH")
	if !ok {
		return fmt.Errorf("%s no establecida", "PROJECTPATH")
	}
	fmt.Printf("%s=%s\n", "PROJECTPATH", projectPath)
	envFilePath, err = filepath.Abs(filepath.Join(projectPath, "./.env"))
	if err != nil {
		return fmt.Errorf("Error al obtener la ruta del archivo .env: %v", err)
	}

	fmt.Println("envFilePath", envFilePath)
	err = godotenv.Load(envFilePath)
	if err != nil {
		return fmt.Errorf("Error al leer el archivo .env: %v", err)
	}
	return nil
}

// ReadVars obtiene el valor de las variables de ambiente
func ReadVars() error {
	err := envconfig.Process("go_crawler", &crawlerEnvVars)
	if err != nil {
		return fmt.Errorf("Error al leer las variables: %v", err)
	}
	return nil
}

func WriteVars() error {
	mapVars := toMap()
	logging.InfoLogger.Printf(".env filepath: %s", envFilePath)
	err := godotenv.Write(mapVars, envFilePath)
	if err != nil {
		return fmt.Errorf("Error al escribir el archivo .env: %v", err)
	}
	return nil
}

// GetEnvs devuelve el valor de la variable de ambiente especificada en envar
func GetEnvs(envar string) (interface{}, error) {
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

func SetEnv(envar string, val interface{}) error {
	switch envar {
	case SeedURL:
		crawlerEnvVars.SeedURL = val.(string)
	default:
		return fmt.Errorf("No se puede modificar el valor de %s", envar)
	}
	return nil
}

func toMap() map[string]string {
	v := reflect.ValueOf(crawlerEnvVars)
	typeOfS := v.Type()
	vars := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		name := typeOfS.Field(i).Name
		vars[name] = fmt.Sprint(v.Field(i).Interface())
	}
	return vars
}
