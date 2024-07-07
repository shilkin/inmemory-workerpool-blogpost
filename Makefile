FILES_GO = $(shell find . -type f -name '*.go' \
	-not -path './vendor/*' \
	-not -path '*/mock/*.go' \
	-not -path '*/mocks/*.go' \
	-not -path '*_mock.go' \
	-not -path '*mock_*.go')

format: ## Format golang code
	@echo "+ $@"
	@go run golang.org/x/tools/cmd/goimports -local "github.com/shilkin" -w $(FILES_GO)
	@go run github.com/daixiang0/gci write \
		-s standard \
		-s default \
		-s "Prefix(github.com/shilkin)" \
		-s "Prefix(github.com/shilkin/inmemory-workerpool-blogpost))" $(FILES_GO)
	@go run mvdan.cc/gofumpt -l -w .
.PHONY: format

lint: ## Run linter
	@echo "+ $@"
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run --allow-parallel-runners
.PHONY: lint

test: ## Run tests
	@echo "+ $@"
	go test ./... -race
.PHONY: lint