GO ?= go
GOLANGCI_LINT ?= golangci-lint
BUILD_DIR ?= bin
COMPOSE ?= docker compose -f deployments/compose.yml
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
TEST_PROJECT ?= $(shell printf '%s' '$(or $(BRANCH),code-winch-test)' | tr 'A-Z' 'a-z' | tr -c 'a-z0-9_-' '-')
WINCH_TEST_PROJECT := $(TEST_PROJECT)
export WINCH_TEST_PROJECT
COMPOSE_TEST ?= $(COMPOSE) -f deployments/compose.test.yml -p $(TEST_PROJECT)
COMPOSE_MAIN ?= $(COMPOSE) -f deployments/compose.main.yml
PROXY_OVERRIDE := $(shell bash scripts/resolve-override.sh deployments/compose.proxy.override.yml)
COMPOSE_PROXY ?= docker compose -f deployments/compose.proxy.yml $(if $(PROXY_OVERRIDE),-f $(PROXY_OVERRIDE))
PROXY_PORT ?= 7090
export PROXY_PORT
HOST_TEST_PORT ?= 7081
HOST_TEST_DIR ?= /tmp/code-winch-host-$(shell printf '%s' '$(CURDIR)' | cksum | cut -d ' ' -f 1)
HOST_TEST_PID := $(HOST_TEST_DIR)/sandbox.pid
HOST_TEST_LOG := $(HOST_TEST_DIR)/sandbox.log
WINCH_E2E_URL ?= http://127.0.0.1:$(HOST_TEST_PORT)
GO_PACKAGES := ./cmd/... ./internal/... ./test/contract/...
# e2e needs a running host or Docker test environment, so it is vetted and
# linted here but only run by targets that expect one to be available.
CHECKED_PACKAGES := $(GO_PACKAGES) ./test/e2e/...
E2E_TEST := go test -count=1 -timeout=10m ./test/e2e/...
# The web gate runs in a node container because the Go toolchain image has no
# node. Order mirrors .github/workflows/web.yml.
WEB_TEST := npm ci --no-audit --no-fund && npm run format:check && npm run lint && npm run typecheck && npm test && npm run build
GO_CHECKS := format-check vet lint test build e2e

.PHONY: all build check docker-e2e e2e format format-check host-test-env host-test-env-down lint main-down main-up proxy proxy-down run test test-cycle test-env test-env-down toolchain-image vet web-build web-test
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
check:
	$(MAKE) $(GO_CHECKS)

e2e:
	WINCH_E2E_URL=$(WINCH_E2E_URL) $(E2E_TEST)
docker-e2e:
	$(COMPOSE_TEST) --profile test run --rm toolchain $(E2E_TEST)
web-test:
	$(COMPOSE_TEST) --profile test run --rm web sh -c '$(WEB_TEST)'
toolchain-image:
	$(COMPOSE_TEST) --profile test build toolchain
proxy:
	$(COMPOSE_PROXY) up -d --wait --no-recreate
	@$(COMPOSE_PROXY) config | grep -q 'host_ip: 127.0.0.1' || { echo "proxy is not published on 127.0.0.1 only" >&2; exit 1; }
proxy-down:
	$(COMPOSE_PROXY) down --remove-orphans
main-up: proxy
	$(COMPOSE_MAIN) up --build -d --wait sandbox
main-down:
	$(COMPOSE_MAIN) down --remove-orphans
test-env: proxy
	$(COMPOSE_TEST) up --build -d --wait sandbox
	@test "$$($(COMPOSE_TEST) exec -T sandbox id -u)" != 0 || { echo "sandbox runs as root" >&2; exit 1; }
	@$(COMPOSE) config | grep -q 'host_ip: 127.0.0.1' || { echo "sandbox is not published on 127.0.0.1 only" >&2; exit 1; }
test-env-down:
	$(COMPOSE_TEST) --profile test down --remove-orphans
host-test-env:
	@set -eu; \
	mkdir -p '$(HOST_TEST_DIR)'; \
	if test -f '$(HOST_TEST_PID)' && kill -0 "$$(cat '$(HOST_TEST_PID)')" 2>/dev/null && \
		ps -o args= -p "$$(cat '$(HOST_TEST_PID)')" 2>/dev/null | grep -Fq '$(CURDIR)/$(BUILD_DIR)/winch-sandbox' && \
		! ps -o stat= -p "$$(cat '$(HOST_TEST_PID)')" 2>/dev/null | grep -q '^Z'; then \
		echo "host test sandbox is already running at $(WINCH_E2E_URL)"; \
		exit 0; \
	fi; \
	rm -f '$(HOST_TEST_PID)'; \
	(cd web && npm ci --no-audit --no-fund && npm run build); \
	$(MAKE) build; \
	WINCH_SANDBOX_ADDR=127.0.0.1:$(HOST_TEST_PORT) WINCH_SANDBOX_STATIC_DIR='$(CURDIR)/web/dist' \
		nohup '$(CURDIR)/$(BUILD_DIR)/winch-sandbox' >'$(HOST_TEST_LOG)' 2>&1 & \
	pid=$$!; echo "$$pid" >'$(HOST_TEST_PID)'; \
	ready=0; \
	for attempt in $$(seq 1 30); do \
		if '$(CURDIR)/$(BUILD_DIR)/winch' status --url '$(WINCH_E2E_URL)' >/dev/null 2>&1; then ready=1; break; fi; \
		if ! kill -0 "$$pid" 2>/dev/null; then break; fi; \
		sleep 1; \
	done; \
	if test "$$ready" != 1; then \
		kill "$$pid" 2>/dev/null || true; rm -f '$(HOST_TEST_PID)'; \
		echo "host test sandbox failed to start; see $(HOST_TEST_LOG)" >&2; exit 1; \
	fi; \
	echo "host test sandbox is running at $(WINCH_E2E_URL) (log: $(HOST_TEST_LOG))"
host-test-env-down:
	@set -eu; \
	if ! test -f '$(HOST_TEST_PID)'; then echo "host test sandbox is not running"; exit 0; fi; \
	pid=$$(cat '$(HOST_TEST_PID)'); \
	if kill -0 "$$pid" 2>/dev/null && ps -o args= -p "$$pid" 2>/dev/null | grep -Fq '$(CURDIR)/$(BUILD_DIR)/winch-sandbox'; then \
		kill "$$pid"; \
		for attempt in $$(seq 1 30); do \
			if ! kill -0 "$$pid" 2>/dev/null || ps -o stat= -p "$$pid" 2>/dev/null | grep -q '^Z'; then break; fi; \
			sleep 1; \
		done; \
		if kill -0 "$$pid" 2>/dev/null && ! ps -o stat= -p "$$pid" 2>/dev/null | grep -q '^Z'; then \
			echo "host test sandbox did not stop" >&2; exit 1; \
		fi; \
	fi; \
	rm -f '$(HOST_TEST_PID)'; \
	echo "host test sandbox stopped"
test-cycle: toolchain-image
	@status=0; $(MAKE) test-env && $(COMPOSE_TEST) --profile test run --rm toolchain make $(GO_CHECKS) || status=$$?; $(MAKE) test-env-down; exit $$status
