VERSION=0.2.16
NAMESPACE=quay.io/nicholasjackson

deps:
	go get github.com/goreleaser/goreleaser

test:
	GOMAXPROCS=7 go test -parallel 7 -cover -race ./...

build:
	go build -o faas-nomad .

build_all:
	goreleaser -snapshot -rm-dist -skip-validate
