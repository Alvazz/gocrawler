package scraper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/leosykes117/gocrawler/pkg/logging"
)

//env es un map que contiene las variable de ambiente del archivo .env
type enVars map[string]string

func ReadEnVars() (enVars, string, error) {
	var err error
	projectPath, ok := os.LookupEnv("PROJECTPATH")
	if !ok {
		return nil, "", fmt.Errorf("%s not set\n", "PROJECTPATH")
	} else {
		logging.InfoLogger.Printf("%s=%s\n", "PROJECTPATH", projectPath)
	}
	envFile, err := filepath.Abs(filepath.Join(projectPath, "./.env"))
	if err != nil {
		return nil, "", fmt.Errorf("Error al obtener l ruta del archivo .env: %v", err)
	}
	vars, err := godotenv.Read(envFile)
	if err != nil {
		return nil, "", fmt.Errorf("Error al leer .env: %v", err)
	}
	logging.InfoLogger.Println("Archivo .env leido correctamente")
	return vars, envFile, nil
}
