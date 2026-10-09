GO ?= go
GOLANGCI_LINT ?= golangci-lint
BUILD_DIR ?= bin
COMPOSE ?= docker compose -f deployments/compose.yml
PROXY_PORT ?= 8088
COMPOSE_PROXY ?= WINCH_PROXY_PORT=$(PROXY_PORT) docker compose -f deployments/compose.proxy.yml
TEST_NAME ?= $(shell (git symbolic-ref --short -q HEAD || git rev-parse --short HEAD) | tr 'A-Z/' 'a-z-' | tr -c 'a-z0-9_\n-' '-')
TEST_PROJECT = code-winch-test-$(TEST_NAME)
COMPOSE_TEST ?= $(COMPOSE) -p $(TEST_PROJECT) -f deployments/compose.routed.yml -f deployments/compose.test.yml
GO_PACKAGES := ./cmd/... ./internal/... ./test/contract/...
# e2e needs a running sandbox, so it is vetted and linted here but only run by
# `make e2e`, `make docker-e2e` and `make test-cycle`.
CHECKED_PACKAGES := $(GO_PACKAGES) ./test/e2e/...
E2E_TEST := go test -count=1 -timeout=10m ./test/e2e/...
# The web gate runs in a node container because the Go toolchain image has no
# node. Order mirrors .github/workflows/web.yml.
WEB_TEST := npm ci --no-audit --no-fund && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build

.PHONY: all build check docker-e2e e2e format format-check lint proxy-down proxy-up run test test-cycle test-env test-env-down toolchain-image vet web-build web-test
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
web-test:
	$(COMPOSE_TEST) --profile test run --rm web sh -c '$(WEB_TEST)'
toolchain-image:
	$(COMPOSE_TEST) --profile test build toolchain
proxy-up:
	$(COMPOSE_PROXY) up -d --wait proxy
proxy-down:
	$(COMPOSE_PROXY) down --remove-orphans
test-env:
	$(COMPOSE_TEST) up --build -d --wait sandbox
	@test "$$($(COMPOSE_TEST) exec -T sandbox id -u)" != 0 || { echo "sandbox runs as root" >&2; exit 1; }
	@$(COMPOSE) config | grep -q 'host_ip: 127.0.0.1' || { echo "sandbox is not published on 127.0.0.1 only" >&2; exit 1; }
	@echo "sandbox: http://127.0.0.1:$(PROXY_PORT)/test/$(TEST_NAME)  [project: $(TEST_PROJECT)]"
	@echo "page:    http://127.0.0.1:$(PROXY_PORT)/test/$(TEST_NAME)/app"
	@echo "(reachable from the host only while the proxy is up: make proxy-up)"
test-env-down:
	$(COMPOSE_TEST) --profile test down --remove-orphans
test-cycle: toolchain-image
	@status=0; $(MAKE) test-env && $(COMPOSE_TEST) --profile test run --rm toolchain sh -c 'make format-check vet test build && $(E2E_TEST)' || status=$$?; $(MAKE) test-env-down; exit $$status
