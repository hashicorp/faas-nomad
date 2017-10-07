VERSION=0.2.2

test:
	GOMAXPROCS=7 go test -parallel 7 -cover -race ./...

build:

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o faas-nomad .

build_docker: build_linux
	docker build -t quay.io/nicholasjackson/faas-nomad:${VERSION} .
