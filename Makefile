VERSION=0.2.16
NAMESPACE=quay.io/nicholasjackson

deps:
	go get github.com/goreleaser/goreleaser
	go get github.com/wadey/gocovmerge

test:
	GOMAXPROCS=7 go test -parallel 7 -cover -race ./...

cover:
	go test -coverprofile=handlers.out ./handlers
	go test -coverprofile=consul.out ./consul
	gocovmerge consul.out handlers.out > c.out

build:
	go build -o faas-nomad .

build_all:
	goreleaser -snapshot -rm-dist
