all: lint test bench

lint:
	golangci-lint run --timeout 5m

test:
	go test -covermode=atomic -race ./...

bench:
	go test -bench=. -benchmem ./...

.PHONY: test lint bench
.DEFAULT_GOAL := all
