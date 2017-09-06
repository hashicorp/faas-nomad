VERSION=0.1

test:
	go test -v -race $(shell go list ./... | grep -v /vendor/)

build:

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o faas-nomad .

build_docker: build_linux
	docker build -t quay.io/nicholasjackson/faas-nomad:${VERSION} .
