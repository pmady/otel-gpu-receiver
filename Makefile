.PHONY: build test lint fmt clean

build:
	go build ./...

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w -s .
	goimports -w .

clean:
	go clean ./...

tidy:
	go mod tidy

.DEFAULT_GOAL := build
