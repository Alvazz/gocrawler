package main

import (
	"fmt"
	"time"

	"github.com/hako/durafmt"
	"github.com/leosykes117/gocrawler/pkg/api"
)

type ResponseAPI struct {
	Success bool        `json:"success"`
	Status  int         `json:"status,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func main() {
	//lambda.Start(HandleRequest)
	start := time.Now()
	HandleRequest()
	elapsed := time.Since(start)
	fmt.Println("Tiempo:", durafmt.Parse(elapsed))
}

func HandleRequest() (string, error) {
	srvc, err := api.NewServices()
	if err != nil {
		return "", err
	}
	if err = srvc.ParserItems(); err != nil {
		return "", err
	}
	return "ejecución correcta", nil
}
