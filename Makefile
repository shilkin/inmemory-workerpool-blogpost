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

example-start: ## Start example of goroutine leak
	docker-compose up -d --remove-orphans
.PHONY: example-start

example-stop: ## Stop example of goroutine leak
	docker-compose down --remove-orphans
.PHONY: example-stop

example-refresh: ## Build and start example of goroutine leak
	docker-compose up -d --build
.PHONY: example-refresh

example-load:  ## Build and start example loadgen
	docker-compose up -d loadgen --build
.PHONY: example-load

example-restart: example-stop example-start ## Restart example of goroutine leak
.PHONY: example-restart

help:
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' | column -c2 -t -s :)"
.PHONY: help

# --go_opt=Mproto/notification/v3/models=./gen/go/proto/notification/v3/models <- go_package
# find proto -type f -name '*.proto' -exec protoc -I=$PWD -I$PWD/proto --go_out=$PWD/gen/go --go_opt=paths=source_relative --go_opt=Mproto/notification/v3/models=./gen/go/proto/notification/v3/models {} \;
twirp:
	protoc --go_out=. \
	--go_opt=Mexample/goroutine-leak/cmd/webserver/proto/user/v1/user.proto=example/goroutine-leak/cmd/webserver/gen/twirp/user/v1 \
	--twirp_out=. ./example/goroutine-leak/cmd/webserver/proto/user/v1/user.proto

connect:
	protoc -I . --go_out=. --connect-go_out=. example/goroutine-leak/cmd/webserver/adapters/connect/connect.proto