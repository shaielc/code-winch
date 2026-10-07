GO ?= go
GOLANGCI_LINT ?= golangci-lint
BUILD_DIR ?= bin
COMPOSE ?= docker compose -f deployments/compose.yml
GO_PACKAGES := ./cmd/... ./internal/... ./test/contract/...
ALL_GO_PACKAGES := ./cmd/... ./internal/... ./test/...

.PHONY: all build check e2e format format-check lint run test test-cycle toolchain-image test-env-down vet web-build
all: check

format:
	$(GO) fmt ./...
format-check:
	@files="$$(gofmt -l $$(find cmd internal test -name '*.go' -type f))"; test -z "$$files" || { echo "Not gofmt-formatted:" >&2; echo "$$files" >&2; exit 1; }
vet:
	$(GO) vet $(ALL_GO_PACKAGES)
lint:
	$(GOLANGCI_LINT) run ./...
test:
	$(GO) test $(GO_PACKAGES)
	$(GO) test -run '^$$' ./test/e2e/...
build:
	mkdir -p "$(BUILD_DIR)"
	$(GO) build -o "$(BUILD_DIR)/winch-sandbox" ./cmd/winch-sandbox
	$(GO) build -o "$(BUILD_DIR)/winch" ./cmd/winch
	$(GO) build -o "$(BUILD_DIR)/fake-harness" ./cmd/fake-harness
run:
	$(GO) run ./cmd/winch-sandbox
web-build:
	cd web && npm install --no-audit --no-fund && npm run build
check: format-check vet lint test build

e2e:
	$(GO) test -count=1 -timeout=10m ./test/e2e/...
toolchain-image:
	$(COMPOSE) --profile test build toolchain
test-env-down:
	$(COMPOSE) --profile test down --remove-orphans
test-cycle: toolchain-image
	@status=0; $(COMPOSE) --profile test run --rm toolchain sh -c 'make format-check vet test build' || status=$$?; $(MAKE) test-env-down; exit $$status
