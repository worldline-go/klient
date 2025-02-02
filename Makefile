.DEFAULT_GOAL := help

.PHONY: lint
lint: ## Lint Go files
	@GOPATH="$(shell dirname $(PWD))" golangci-lint run ./...

.PHONY: test
test: ## Run unit tests
	@go test -v -race ./...

.PHONY: coverage
coverage: ## Run unit tests with coverage
	@go test -v -race -cover -coverpkg=./... -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
