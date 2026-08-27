GO ?= go
GOLANGCI_LINT ?= golangci-lint

COMPOSE ?= docker compose -f deployments/compose.yml
IN_RUNNER = $(COMPOSE) exec -T runner
TEST_DATABASE ?= winch_test
GO_PACKAGES := ./cmd/... ./internal/... ./pkg/... ./test/...

# Targets are marked [host] or [docker].
#
#   [host]   runs directly on this machine and needs go, node, or
#            golangci-lint installed. CI uses these.
#   [docker] needs Docker only; the `runner` container supplies the Go
#            toolchain and postgres supplies the database.
#
# With no Go toolchain installed, the [docker] group is the way in. See
# deployments/README.md for the testing procedure.

.PHONY: all api-check api-compat api-generate api-validate build check format \
	format-check lint run runner-image runner-integration runner-shell runner-verify \
	test test-cycle test-env test-env-down test-integration vet web-build

all: check

# ---------------------------------------------------------------- host targets

## api-generate: [host] Regenerate server and browser types from the OpenAPI source.
api-generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 --config api/openapi/oapi-codegen.yaml api/openapi/code-winch.yaml
	cd web && npm run api:generate

## api-validate: [host] Parse and validate the OpenAPI document and contract fixtures.
api-validate:
	go test ./test/contract/openapi -run TestOpenAPIContract

## api-compat: [host] Reject breaking changes relative to the committed v1 baseline.
api-compat:
	go run github.com/oasdiff/oasdiff@v1.11.7 breaking test/contract/openapi/v1.yaml api/openapi/code-winch.yaml --fail-on ERR

## api-check: [host] Verify validation, compatibility, and deterministic generated output.
api-check: api-validate api-compat
	@tmp="$$(mktemp -d)"; trap 'rm -rf "$$tmp"' EXIT; \
	cp internal/adapters/transport/httpapi/types.gen.go "$$tmp/go"; \
	cp web/src/api/schema.ts "$$tmp/ts"; \
	$(MAKE) api-generate; \
	cmp "$$tmp/go" internal/adapters/transport/httpapi/types.gen.go; \
	cmp "$$tmp/ts" web/src/api/schema.ts

## format: [host] Format all Go source files.
format:
	$(GO) fmt ./...

## format-check: [host] Fail if any Go source file is not gofmt-formatted.
format-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "The following Go files are not formatted:" >&2; \
		echo "$$files" >&2; \
		echo "Run 'make format' to fix them." >&2; \
		exit 1; \
	fi

## vet: [host] Report suspicious Go constructs.
vet:
	$(GO) vet ./...

## lint: [host] Run the pinned repository-wide Go linter.
lint:
	$(GOLANGCI_LINT) run

## test: [host] Run all Go unit tests.
test:
	$(GO) test $(GO_PACKAGES)

## test-integration: [host] Run PostgreSQL integration tests.
## Requires PG_TEST_DATABASE_URL to name a disposable database.
test-integration:
	$(GO) test -tags integration ./internal/adapters/postgres/...

## build: [host] Build the daemon.
build:
	@output="$$(mktemp)"; \
	trap 'rm -f "$$output"' EXIT; \
	$(GO) build -o "$$output" ./cmd/winchd

## run: [host] Start the daemon with configuration resolved from file and environment.
## Serves the API alone unless web-build has produced web/dist.
run:
	$(GO) run ./cmd/winchd

## web-build: [host] Build the browser assets the daemon serves from WINCH_STATIC_DIR.
web-build:
	cd web && npm install --no-audit --no-fund && npm run build

## check: [host] Run every check required by CI.
check: api-check format-check vet lint test build

# ------------------------------------------------------------- docker targets

## runner-image: [docker] Build the toolchain image the runner and daemon share.
## Needs registry access; test-env builds on demand if the image is missing.
runner-image:
	$(COMPOSE) --profile test build

## test-env: [docker] Start the runner and postgres, and create the test database.
test-env:
	$(COMPOSE) --profile test up -d
	@$(COMPOSE) exec -T postgres psql -U winch -d winch -tAc \
		"SELECT 1 FROM pg_database WHERE datname = '$(TEST_DATABASE)'" \
		| grep -q 1 \
		|| $(COMPOSE) exec -T postgres createdb -U winch $(TEST_DATABASE)

## runner-verify: [docker] Format, vet, unit-test, and compile everything in the runner.
## The container has no golangci-lint or npm, so this is `check` minus lint and api-check.
runner-verify: test-env
	@$(IN_RUNNER) sh -c 'files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "Not gofmt-formatted:" >&2; echo "$$files" >&2; exit 1; fi'
	$(IN_RUNNER) go vet ./...
	$(IN_RUNNER) go test ./...
	$(IN_RUNNER) go build ./...

## runner-integration: [docker] Run the build-tagged integration suite in the runner.
runner-integration: test-env
	$(IN_RUNNER) go test -tags integration ./...

## test-env-down: [docker] Stop and remove the runner and drop the test database,
## leaving the daemon and its own database running.
test-env-down:
	$(COMPOSE) --profile test stop runner
	$(COMPOSE) --profile test rm -f runner
	@if [ "$(TEST_DATABASE)" = "winch" ]; then \
		echo "refusing to drop '$(TEST_DATABASE)': that is the daemon's database" >&2; \
		exit 1; \
	fi; \
	$(COMPOSE) exec -T postgres dropdb -U winch --if-exists --force "$(TEST_DATABASE)" \
		|| echo "note: kept '$(TEST_DATABASE)'; postgres is not running" >&2

## test-cycle: [docker] Build, start, verify, integration-test, and tear down.
## Tears down even when a step fails, and exits with that step's status.
test-cycle: runner-image test-env
	@status=0; \
	$(MAKE) runner-verify runner-integration || status=$$?; \
	$(MAKE) test-env-down; \
	exit $$status

## runner-shell: [docker] Open an interactive shell in the runner.
runner-shell: test-env
	$(COMPOSE) exec runner bash
