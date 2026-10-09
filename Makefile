GO ?= go
GOLANGCI_LINT ?= golangci-lint
BUILD_DIR ?= bin
COMPOSE ?= docker compose -f deployments/compose.yml
COMPOSE_TEST ?= $(COMPOSE) -f deployments/compose.test.yml
GO_PACKAGES := ./cmd/... ./internal/... ./test/contract/...
# e2e needs a running sandbox, so it is vetted and linted here but only run by
# `make e2e`, `make docker-e2e` and `make test-cycle`.
CHECKED_PACKAGES := $(GO_PACKAGES) ./test/e2e/...
E2E_TEST := go test -count=1 -timeout=10m ./test/e2e/...

.PHONY: all build check docker-e2e e2e format format-check lint run test test-cycle test-env test-env-down toolchain-image vet web-build
all: check

format:
	$(GO) fmt ./...
format-check:
	@files="$$(gofmt -l $$(find cmd internal test -name '*.go' -type f))"; test -z "$$files" || { echo "Not gofmt-formatted:" >&2; echo "$$files" >&2; exit 1; }
vet:
	$(GO) vet $(CHECKED_PACKAGES)
lint:
	$(GOLANGCI_LINT) run $(CHECKED_PACKAGES)
test:
	$(GO) test $(GO_PACKAGES)
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
	$(E2E_TEST)
docker-e2e:
	$(COMPOSE_TEST) --profile test run --rm toolchain $(E2E_TEST)
toolchain-image:
	$(COMPOSE_TEST) --profile test build toolchain
test-env:
	$(COMPOSE_TEST) up --build -d --wait sandbox
	@test "$$($(COMPOSE_TEST) exec -T sandbox id -u)" != 0 || { echo "sandbox runs as root" >&2; exit 1; }
	@$(COMPOSE) config | grep -q 'host_ip: 127.0.0.1' || { echo "sandbox is not published on 127.0.0.1 only" >&2; exit 1; }
	@echo "sandbox: http://$$($(COMPOSE_TEST) port sandbox 8080)"
test-env-down:
	$(COMPOSE_TEST) --profile test down --remove-orphans
test-cycle: toolchain-image
	@status=0; $(MAKE) test-env && $(COMPOSE_TEST) --profile test run --rm toolchain sh -c 'make format-check vet test build && $(E2E_TEST)' || status=$$?; $(MAKE) test-env-down; exit $$status
