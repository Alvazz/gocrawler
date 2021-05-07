package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/leosykes117/gocrawler/pkg/itemparser"
)

type ResponseAPI struct {
	Success bool        `json:"success"`
	Status  int         `json:"status,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func main() {
	lambda.Start(noAPI)
}

func _(ev events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET,HEAD,OPTIONS,POST",
		},
		StatusCode: http.StatusInternalServerError,
	}

	result, err := itemparser.GetItemsFromCache()
	if err != nil {
		return resp, err
	}

	bytes, err := json.Marshal(ResponseAPI{
		Success: true,
		Status:  http.StatusOK,
		Result:  result,
	})
	if err != nil {
		return resp, err
	}

	resp.Body = string(bytes)
	resp.StatusCode = http.StatusOK

	return resp, nil
}

func noAPI() (string, error) {
	result, err := itemparser.GetItemsFromCache()
	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(ResponseAPI{
		Success: true,
		Status:  http.StatusOK,
		Result:  result,
	})
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
