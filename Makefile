#!/usr/bin/env make
.DEFAULT_GOAL := help
.SHELLFLAGS   := -eou pipefail

.PHONY: build
build: ## Build package.
	@go build .

.PHONY: test
test: ## Run tests.
	@go clean -testcache && go test -race ./...

.PHONY: help
help: ## Show help/usage.
	@grep -E '^[%a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
