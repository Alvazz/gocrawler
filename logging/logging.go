package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func InitLogging() {
	file, err := createFile()
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func createFile() (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Ocurrio un error al obtener la ruta del directorio home: %v", err)
	}

	dir, err := filepath.Abs(filepath.Join(home, "./crawling-data/logs/"))
	if err != nil {
		return nil, fmt.Errorf("Ocurrio un error al obtener la ruta del directorio de los logs: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(err)
			fmt.Println("Creando directorio de logs")
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return nil, fmt.Errorf("Ocurrio un error al crear el directorio de los logs: %v", err)
			}
		} else {
			return nil, err
		}
	}

	logFileName := time.Now().Format("060102_150405") + ".log"
	logFile := filepath.Join(dir, logFileName)
	fmt.Printf("Log file in: %s\n", logFile)
	if err != nil {
		return nil, fmt.Errorf("Ocurrio un error al formar la ruta para el archivo log: %v", err)
	}
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("Ocurrio un error al abrir el archivo del log: %v", err)
	}
	return file, nil
}
