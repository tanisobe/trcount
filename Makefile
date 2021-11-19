BINDIR:=bin
REVISION:=$(shell git rev-parse --short HEAD)
TAG:=$(shell git describe --tags)

.PHONY: build
build:
	go build -o bin/trcount -ldflags " -X main.revision=$(REVISION)" cmd/trcount/main.go

.PHONY: test
test:
	cd test && go test


.PHONY: docker
docker:
	docker build -t trcount/test:latest deployments
	docker run --detach trcount/test:latest

.PHONY: clean
clean:
	-@rm *.log && rm /bin/trcount


