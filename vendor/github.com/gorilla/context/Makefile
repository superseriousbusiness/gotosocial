GO_LINT=$(shell which golangci-lint 2> /dev/null || echo '')
GO_LINT_URI=github.com/golangci/golangci-lint/cmd/golangci-lint@latest

GO_SEC=$(shell which gosec 2> /dev/null || echo '')
GO_SEC_URI=github.com/securego/gosec/v2/cmd/gosec@latest

GO_VULNCHECK=$(shell which govulncheck 2> /dev/null || echo '')
GO_VULNCHECK_URI=golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: golangci-lint
golangci-lint: ## Run golangci-lint. Example: make golangci-lint
	$(if $(GO_LINT), ,go install $(GO_LINT_URI))
	@echo "##### Running golangci-lint #####"
	golangci-lint run -v

.PHONY: verify
verify: ## Run all verifications [golangci-lint]. Example: make verify
	@echo "##### Running verifications #####"
	$(MAKE) golangci-lint

.PHONY: gosec
gosec: ## Run gosec. Example: make gosec
	$(if $(GO_SEC), ,go install $(GO_SEC_URI))
	@echo "##### Running gosec #####"
	gosec ./...

.PHONY: govulncheck
govulncheck: ## Run govulncheck. Example: make govulncheck
	$(if $(GO_VULNCHECK), ,go install $(GO_VULNCHECK_URI))
	@echo "##### Running govulncheck #####"
	govulncheck ./...

.PHONY: security
security: ## Run all security checks [gosec, govulncheck]. Example: make security
	@echo "##### Running security checks #####"
	$(MAKE) gosec
	$(MAKE) govulncheck

.PHONY: test-unit
test-unit: ## Run unit tests. Example: make test-unit
	@echo "##### Running unit tests #####"
	go test -race -cover -coverprofile=coverage.coverprofile -covermode=atomic -v ./...

.PHONY: test
test: ## Run all tests [test-unit]. Example: make test
	@echo "##### Running tests #####"
	$(MAKE) test-unit

.PHONY: help
help: ## Print this help. Example: make help
	@echo "##### Printing help #####"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
