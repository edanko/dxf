GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_VERSION ?= v2.15.0
DXF_EZDXF_SOURCE ?= ezdxf==1.4.5
COVERAGE_FILE ?= coverage.out

.DEFAULT_GOAL := help
.PHONY: help build test lint lint-install fmt fmt-check vet check crosscheck cover tidy clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-14s %s\n", $$1, $$2}'

build: ## Build all packages
	$(GO) build ./...

test: ## Run all tests
	$(GO) test ./...

lint: ## Run golangci-lint (v2)
	$(GOLANGCI_LINT) run ./...

lint-install: ## Install the pinned golangci-lint v2 release
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

fmt: ## Format code with goimports via golangci-lint
	$(GOLANGCI_LINT) fmt ./...

fmt-check: ## Check formatting without modifying files
	$(GOLANGCI_LINT) fmt --diff ./...

vet: ## Run go vet
	$(GO) vet ./...

check: test lint ## Run tests and lint

crosscheck: ## Run opt-in ezdxf semantic cross-check
	DXF_EZDXF_CROSSCHECK=1 DXF_EZDXF_SOURCE="$(DXF_EZDXF_SOURCE)" \
		$(GO) test -count=1 -run TestEzdxfCrossCheck .

cover: ## Run tests with a coverage profile
	$(GO) test -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

clean: ## Remove build and coverage artifacts
	$(GO) clean ./...
	rm -f $(COVERAGE_FILE)
