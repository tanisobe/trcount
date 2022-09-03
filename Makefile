BINDIR:=bin
VERSION:=$(shell git describe --tags --abbrev=0)
REVISION:=$(shell git rev-parse --short HEAD)
TAG:=$(shell git describe --tags)
DOCKER_IMAGE="elpadrinoiv/snmp_mock"
CONTAINER_NAME="snmp_mock"
SNMP_COMMUNITY="my_comm"
DUMMY_OIDS_DIR="/testdata/sample_oids"
.PHONY: build
build:
	go build -o bin/trmon -ldflags " -X main.version=$(VERSION) -X main.revision=$(REVISION)" cmd/trmon/main.go

.PHONY: test
test: docker
	go test -v -cover
	-@docker kill $(CONTAINER_NAME)

.PHONY: docker
docker:
	docker run -d -p 127.0.0.1:161:161/udp --name $(CONTAINER_NAME) --rm \
	-e SNMP_COMMUNITY=$(SNMP_COMMUNITY) \
	-v $(DUMMY_OIDS_DIR):/app/oids $(DOCKER_IMAGE)

.PHONY: clean
clean:
	-@rm *.log && rm /bin/trmon
	-@docker kill $(CONTAINER_NAME)
