ARCH:=$(shell uname -m)
IMG ?= ruijzhan/chinadns-go:$(ARCH)

all: build

fmt:
	go fmt ./...

test:
	go vet ./...
	go test ./...

build:
	CGO_ENABLED=0 go build ./cmd/chinadns
	CGO_ENABLED=0 go build ./cmd/chnroutegen

docker-build: test
	docker build . -t $(IMG)

docker-push:
	docker push $(IMG)

chnroute:
	./chnroutegen
	echo "chnroute.json generated"

clean:
	rm -f chinadns chnroutegen