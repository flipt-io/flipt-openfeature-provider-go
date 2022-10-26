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

lint:
	@echo "Linting..."
	@golangci-lint run

test:
	@echo "Testing..."
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...

cover:
	@echo "Testing with coverage..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out

doc:
	@echo "Generating documentation..."
	@echo "	http://localhost:6060/pkg/github.com/flipt-io/openfeature-provider-go/"
	@godoc -http=:6060 -goroot .

.DEFAULT_GOAL := build