.PHONY: all dep build clean test lint

all: lint build test

lint: ## Lint the files
	@cd ./api-server && golangci-lint run ./...
	@cd ./api-client && golangci-lint run ./...

test: ## Run unittests for each module
	@cd ./api-server && go test -short

dep: ## Get the dependencies
	@cd ./api-server && go get -v -d ./...
	@cd ./api-client && go get -v -d ./...


build: dep ## Build the binary file
	@cd ./api-server && go build -v
	@cd ./api-client && go build -v

clean: ## Remove previous build
	@rm -f ./api-server/api-server
	@rm -f ./api-client/api-client

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'