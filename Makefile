VERSION=0.1

build:

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o faas-nomad .

build_docker: build_linux
	docker build -t quay.io/nicholasjackson/faas-nomad:${VERSION} .
