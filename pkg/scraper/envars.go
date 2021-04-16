package scraper

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/leosykes117/gocrawler/pkg/logging"
)

//env es un map que contiene las variable de ambiente del archivo .env
type enVars map[string]string

func ReadEnVars() {
	var err error
	projectPath, ok := os.LookupEnv("PROJECTPATH")
	if !ok {
		logging.ErrorLogger.Fatalf("%s not set\n", "PROJECTPATH")
	} else {
		logging.InfoLogger.Printf("%s=%s\n", "PROJECTPATH", projectPath)
	}
	envFilePath, err = filepath.Abs(filepath.Join(projectPath, "./.env"))
	if err != nil {
		logging.ErrorLogger.Fatalf("Error al obtener l ruta del archivo .env: %v", err)
	}
	crawlerVars, err = godotenv.Read(envFilePath)
	if err != nil {
		logging.ErrorLogger.Fatalf("Error al leer .env: %v", err)
	}
	logging.InfoLogger.Println("Archivo .env leido correctamente")
}
