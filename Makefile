BINARY := canopy
MODULE := github.com/nhomble/canopy
VERSION ?= dev

.PHONY: build test test-e2e clean install

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/canopy

test:
	go test ./...

test-e2e: build
	./scripts/test-e2e.sh

clean:
	rm -f $(BINARY)

install:
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/canopy
