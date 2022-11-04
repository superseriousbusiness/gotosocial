help: ## show help, shown by default if no target is specified
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

lint: ## run code linters
	golangci-lint run

benchmark: ## run benchmarks
	cd benchmarks && perflock go test -cpu 8 -run=^# -bench=.

benchmark-perflock: ## run benchmarks using perflock - https://github.com/aclements/perflock
	cd benchmarks && perflock -governor 80% go test -count 3 -cpu 8 -run=^# -bench=.

test: ## run tests
	go test -race ./...
	GOARCH=386 go test ./...

test-coverage: ## run unit tests and create test coverage
	go test ./... -coverprofile .testCoverage -covermode=atomic -coverpkg=./...

test-coverage-web: test-coverage ## run unit tests and show test coverage in browser
	go tool cover -func .testCoverage | grep total | awk '{print "Total coverage: "$$3}'
	go tool cover -html=.testCoverage

install-linters: ## install all used linters
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.49.0
