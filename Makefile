.PHONY: fmt check-fmt lint vet test

GO_PKGS   := $(shell go list -f {{.Dir}} ./...)

tidy: ## Run go mod tidy in all directories
	go mod tidy
.PHONY: tidy

fmt: format-code
format-code: tidy ## Format go code and run the fixer, alias: fmt
	golangci-lint fmt
	golangci-lint run --fix ./...
.PHONY: fmt format-code

test:
	@go test -race -v $(GO_FLAGS) -count=1 $(GO_PKGS)
.PHONY: t test
