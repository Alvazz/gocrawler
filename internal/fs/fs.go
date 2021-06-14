package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ErrGetUserHomeDir = "Ocurrio un error al obtener el directorio del usuario"
)

func CreateFile(filename string) (string, error) {
	var err error
	if filepath.IsAbs(filename) {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return "", err
		}
	}
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			fmt.Println(err)
			fmt.Printf("Creando directorio de %s\n", dir)
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", fmt.Errorf("Ocurrio un error al crear el directorio de los logs: %v", err)
			}
		} else {
			return "", err
		}
	}
	return filename, err
}

func GetUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ErrGetUserHomeDir
	}
	return home
}

func Abs(file string) string {
	if filepath.IsAbs(file) {
		return file
	}
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return ""
	}
	return fileAbs
}
