.PHONY: build test clean tidy

APP_NAME=gophership
CTL_NAME=gs-ctl
GO_VERSION=1.22

VERSION?=1.0.0
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"

build: tidy build-dashboard
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)
	go build $(LDFLAGS) -o bin/$(CTL_NAME) ./cmd/$(CTL_NAME)

build-all: build-dashboard
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

build-dashboard:
	cd dashboard && npm install && npm run build

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

tidy:
	go mod tidy

clean:
	rm -rf bin/
