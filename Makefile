ARCH:=$(shell uname -m)
IMG ?= ruijzhan/chinadns-go:$(ARCH)

all: build

fmt:
	go fmt ./...

test:
	go vet ./...
	go test ./...

build:
	CGO_ENABLED=0 go build -o chinadns ./cmd

docker-build: test
	docker build . -t $(IMG)

docker-push:
	docker push $(IMG)

clean:
	rm -f chinadns