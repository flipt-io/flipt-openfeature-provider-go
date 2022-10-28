.PHONY: build
build:
	@echo "Building..."
	@go build ./...

.PHONY: fmt
fmt:
	@echo "Formatting..."
	@gofmt -s -w .
	@goimports -w .

.PHONY: generate
generate:
	@echo "Generating..."
	@go generate ./...

.PHONY: lint
lint:
	@echo "Linting..."
	@golangci-lint run

.PHONY: test
test:
	@echo "Testing..."
	@go test --short -v -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: integration-test
integration-test: # dependent on `docker run -p 8080:8080 ghcr.io/flipt-io/flipt-openfeature-testbed:latest`
	@echo "Running integration tests..."
	git submodule update --init --recursive
	go test -v ./...

.PHONY: cover
cover:
	@echo "Testing with coverage..."
	@go test --short -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out

.PHONY: doc
doc:
	@echo "Generating documentation..."
	@echo "	http://localhost:6060/pkg/github.com/flipt-io/openfeature-provider-go/"
	@godoc -http=:6060 -goroot .

.DEFAULT_GOAL := build