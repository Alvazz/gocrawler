package env

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var (
	envFiles map[string]string = map[string]string{
		"go_crawler": "",
		"aws":        "",
	}
)

// LoadVars carga el archivo .env en variables de ambiente
func LoadVars(filenames ...string) error {
	var (
		err          error
		envFilePaths = make([]string, 0)
	)
	projectPath, ok := os.LookupEnv("PROJECTPATH")
	if !ok {
		return fmt.Errorf("%s no establecida", "PROJECTPATH")
	}

	fmt.Printf("%s=%s\n", "PROJECTPATH", projectPath)

	if filenames == nil || len(filenames) < 1 {
		filenames, err = getAllEnvFiles(projectPath)
		if err != nil {
			return err
		}
		envFiles = map[string]string{}
	}

	for _, file := range filenames {
		filePath, err := filepath.Abs(filepath.Join(projectPath, fmt.Sprintf("./.%s.env", file)))
		if err != nil {
			return fmt.Errorf("Error al obtener la ruta del archivo .env: %v", err)
		}
		envFilePaths = append(envFilePaths, filePath)
		envFiles[file] = filePath
	}

	fmt.Println("envFilePath", envFilePaths)
	err = godotenv.Load(envFilePaths...)
	if err != nil {
		return fmt.Errorf("Error al leer el archivo .env: %v", err)
	}
	return nil
}

// ReadVars obtiene el valor de las variables de ambiente
func ReadVars(envarprefixes ...string) error {
	var (
		i interface{}
	)
	if envarprefixes == nil || len(envarprefixes) == 0 {
		for k, _ := range envFiles {
			envarprefixes = append(envarprefixes, k)
		}
	}
	fmt.Println("envarprefixes", envarprefixes)
	for _, prefix := range envarprefixes {
		switch prefix {
		case "go_crawler":
			i = &crawlerEnvVars
		case "aws":
			i = &awsEnvVars
		default:
			return fmt.Errorf("No existe la configuración %s", prefix)
		}
		err := envconfig.Process(prefix, i)
		if err != nil {
			return fmt.Errorf("Error al leer las variables: %v", err)
		}
	}
	return nil
}

func WriteVars(envar ...string) error {
	var (
		mapVars map[string]string
	)
	if envar == nil || len(envar) == 0 {
		for k, _ := range envFiles {
			envar = append(envar, k)
		}
	}
	for _, file := range envar {
		switch file {
		case "go_crawler":
			mapVars = toMap(&crawlerEnvVars)
		case "aws":
			mapVars = toMap(&awsEnvVars)
		default:
			return fmt.Errorf("No existe la configuración %s", file)
		}
		err := godotenv.Write(mapVars, envFiles[file])
		if err != nil {
			return fmt.Errorf("Error al escribir el archivo .env: %v", err)
		}
	}
	return nil
}

func toMap(i interface{}) map[string]string {
	v := reflect.ValueOf(i)
	typeOfS := v.Type()
	vars := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		tag := string(typeOfS.Field(i).Tag)
		name := tag[strings.Index(tag, ":")+2 : len(tag)-1]
		vars[name] = fmt.Sprint(v.Field(i).Interface())
	}
	return vars
}

func getAllEnvFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(info.Name()) == ".env" {
			filename := info.Name()
			filename = filename[1 : len(filename)-len(filepath.Ext(filename))]
			files = append(files, filename)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for i, file := range files {
		fmt.Printf("%d: %s\n", i, file)
	}
	return files, nil
}
