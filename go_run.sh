#!/bin/bash

docker run --rm \
  -v "$PWD":/go/src/gocrawler \
  -w /go/src/gocrawler \
  --name go golang:1.15.6-alpine3.12 \
  go build -v -o bin/gocrawler

cd bin
./gocrawler
