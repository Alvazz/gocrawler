#!/bin/bash

echo Creando el contenedor
docker run -it --rm \
	--mount type=bind,source="$(PWD)"/bin,target=/go-workspace/bin \
	-w /go-workspace/bin \
	--name go-crawler debug-app \
	curl https://httpbin.org/ip
