.PHONY: build

build:
	@echo "Building..."
	@go build ./...

fmt:
	@echo "Formatting..."
	@gofmt -s -w .
	@goimports -w .

generate:
	@echo "Generating..."
	@go generate ./...

test:
	@echo "Testing..."
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...

cover:
	@echo "Testing with coverage..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out

.DEFAULT_GOAL := build