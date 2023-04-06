TEST_TIMEOUT = 20
.DEFAULT_GOAL := help

REQUIRED_BINS := tparse
$(foreach bin,$(REQUIRED_BINS),\
    $(if $(shell command -v $(bin) 2> /dev/null),,$(error Please install `$(bin)`)))

.PHONY: help
help: ## Display usage
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make [target] \033[36m\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)


.PHONY: tidy
tidy: ## Sync and clean up Go dependencies
	@go mod tidy -v -compat=1.17

# intended for local dev use (won't check for unclean git)
.PHONY: build
build: unit-test ## Build binary
	$(info Make: Building binary "$(DIST_FOLDER)/$(BINARY)".)
	@goreleaser release --rm-dist --snapshot

.PHONY: gofmt
gofmt: ## Check code is gofmted
	$(eval PKGS=$(shell go list -f {{.Dir}} ./...))
	@test -z "$(shell gofmt -s -l -d -e ${PKGS} | tee /dev/stderr)"

.PHONY: unit-test
unit-test: gofmt tidy ## Run unit tests
	$(eval TAG=unit)
	@go test -count 1 -tags=${TAG} -race -timeout $(TEST_TIMEOUT)s -cover -v -json ./... | tparse -all -notests

.PHONY: integration-test
integration-test: gofmt ## Run integration tests (requires Kubernetes cluster w/ OpenFaaS or use hack/run_integration_tests.sh)
	$(info Make: Running integration tests.)
	$(eval TAG=integration,openfaas)
	@go test -count 1 -tags=${TAG} -race -timeout $(TEST_TIMEOUT)s -cover -v ./... -ginkgo.v

.PHONY: test
test: unit-test integration-test ## Run unit and integration tests
