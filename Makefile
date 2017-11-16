VERSION=0.2.6
NAMESPACE=quay.io/nicholasjackson

test:
	GOMAXPROCS=7 go test -parallel 7 -cover -race ./...

build:
	go build -o faas-nomad .

build_linux:
	CGO_ENABLED=0 GOOS=linux go build -o faas-nomad .

build_docker: build_linux
	docker build -t ${NAMESPACE}/faas-nomad:${VERSION} .
	docker tag ${NAMESPACE}/faas-nomad:${VERSION} faas-nomad:${VERSION}
	docker tag ${NAMESPACE}/faas-nomad:${VERSION} ${NAMESPACE}/faas-nomad:latest

push_docker: 
	docker push ${NAMESPACE}/faas-nomad:${VERSION}
	docker push ${NAMESPACE}/faas-nomad:latest
