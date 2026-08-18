GO ?= go
GOLANGCI_LINT ?= golangci-lint

OAPI_CODEGEN ?= github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
# One schema file per resource, mirroring api/openapi/paths/. Discovered from
# the directory rather than listed, so adding a resource does not edit this
# file too.
API_COMPONENTS = $(basename $(notdir $(wildcard api/openapi/components/*.yaml)))

COMPOSE ?= docker compose -f deployments/compose.yml
IN_RUNNER = $(COMPOSE) exec -T runner
# The browser toolchain runs from web/, where its package.json lives.
IN_RUNNER_WEB = $(COMPOSE) exec -T -w /src/web runner
TEST_DATABASE ?= winch_test

# Targets are marked [host] or [docker].
#
#   [host]   runs directly on this machine and needs go, node, or
#            golangci-lint installed. CI uses these.
#   [docker] needs Docker only; the `runner` container supplies the Go and
#            Node toolchains and postgres supplies the database.
#
# With no toolchain installed, the [docker] group is the way in. golangci-lint
# is the only tool the container lacks, so `lint` is the only CI gate it cannot
# reproduce. See deployments/README.md for the testing procedure.

.PHONY: all api-check api-check-go api-compat api-generate api-generate-go \
	api-validate build check format format-check lint run runner-image \
	runner-shell runner-verify runner-web-deps runner-web-verify test \
	test-cycle test-env test-env-down test-integration vet web-build

all: check

# ---------------------------------------------------------------- host targets

## api-generate: [host] Regenerate server and browser types from the OpenAPI source.
## Both halves, so it needs npm as well as Go; api-generate-go is the Go half alone.
api-generate: api-generate-go
	cd web && npm run api:generate

## api-generate-go: [host] Regenerate the server types only. Needs no npm, so it
## runs wherever $(GO) does, including the runner container.
## The components pass is not optional: the root document maps its external
## refs to "-", which resolves them in-package without declaring them, so
## without this loop types.gen.go references types nothing defines.
api-generate-go:
	$(GO) run $(OAPI_CODEGEN) --config api/openapi/oapi-codegen.yaml api/openapi/code-winch.yaml
	@for f in $(API_COMPONENTS); do \
		echo "$(GO) run $(OAPI_CODEGEN) --config api/openapi/oapi-codegen-components.yaml api/openapi/components/$$f.yaml"; \
		$(GO) run $(OAPI_CODEGEN) --config api/openapi/oapi-codegen-components.yaml \
			-o internal/adapters/transport/httpapi/$$(echo $$f | tr - _).gen.go \
			api/openapi/components/$$f.yaml || exit 1; \
	done

## api-validate: [host] Parse and validate the OpenAPI document and contract fixtures.
api-validate:
	$(GO) test ./test/contract/openapi -run TestOpenAPIContract

## api-compat: [host] Reject breaking changes relative to the committed v1 baseline.
api-compat:
	$(GO) run github.com/oasdiff/oasdiff@v1.11.7 breaking test/contract/openapi/v1.yaml api/openapi/code-winch.yaml --fail-on ERR

## api-check: [host] Verify validation, compatibility, and deterministic generated output.
## Every *.gen.go is compared, not just types.gen.go, and so is the file list:
## the components split means a regeneration can add or drop a file, which a
## fixed list of names would not notice.
## The steps are chained with && deliberately: with `;` a failed regeneration was
## swallowed and the target reported success having generated nothing.
api-check: api-validate api-compat
	@tmp="$$(mktemp -d)"; trap 'rm -rf "$$tmp"' EXIT; \
	mkdir "$$tmp/go" && cp internal/adapters/transport/httpapi/*.gen.go "$$tmp/go/" && \
	ls internal/adapters/transport/httpapi/*.gen.go > "$$tmp/list" && \
	cp web/src/api/schema.ts "$$tmp/ts" && \
	$(MAKE) api-generate && \
	ls internal/adapters/transport/httpapi/*.gen.go | cmp - "$$tmp/list" && \
	for f in internal/adapters/transport/httpapi/*.gen.go; do \
		cmp "$$tmp/go/$$(basename $$f)" "$$f" || exit 1; \
	done && \
	cmp "$$tmp/ts" web/src/api/schema.ts

## api-check-go: [host] api-check without the browser half, for a host that has Go
## but no Node. It does not cover web/src/api/schema.ts, so it is not a substitute
## for api-check in CI. The runner container carries npm and runs the full
## api-check, so the [docker] path no longer needs this.
api-check-go: api-validate api-compat
	@tmp="$$(mktemp -d)"; trap 'rm -rf "$$tmp"' EXIT; \
	mkdir "$$tmp/go" && cp internal/adapters/transport/httpapi/*.gen.go "$$tmp/go/" && \
	ls internal/adapters/transport/httpapi/*.gen.go > "$$tmp/list" && \
	$(MAKE) api-generate-go GO='$(GO)' && \
	ls internal/adapters/transport/httpapi/*.gen.go | cmp - "$$tmp/list" && \
	for f in internal/adapters/transport/httpapi/*.gen.go; do \
		cmp "$$tmp/go/$$(basename $$f)" "$$f" || exit 1; \
	done

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
	$(GO) test ./...

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
## runner-web-verify builds them inside the container when there is no host Node.
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

## runner-web-deps: [docker] Install web/node_modules inside the runner.
## They live in a named volume rather than the bind-mounted working tree, so the
## container's install never overwrites the host's; the volume outlives
## test-env-down, so repeat runs only reconcile it.
runner-web-deps: test-env
	$(IN_RUNNER_WEB) npm install --no-audit --no-fund

## runner-verify: [docker] Run CI's `check` inside the container, minus lint.
## golangci-lint is the only tool the image lacks; api-check now runs here in
## full, browser types included, because the container carries npm. This defers
## to the [host] targets rather than restating them, so the two paths cannot drift.
runner-verify: test-env runner-web-deps
	$(IN_RUNNER) make api-check format-check vet test build

## runner-web-verify: [docker] Run the browser gates from .github/workflows/web.yml
## in the runner: formatting, lint, types, the vitest suite, and the production
## build. That workflow's api:check is covered by runner-verify's api-check.
runner-web-verify: runner-web-deps
	$(IN_RUNNER_WEB) npm run format:check
	$(IN_RUNNER_WEB) npm run lint
	$(IN_RUNNER_WEB) npm run typecheck
	$(IN_RUNNER_WEB) npm test
	$(IN_RUNNER_WEB) npm run build

## test-integration: [docker] Run the build-tagged integration suite in the runner.
test-integration: test-env
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
## Every CI gate but lint, browser suite included.
## Tears down even when a step fails, and exits with that step's status.
test-cycle: runner-image test-env
	@status=0; \
	$(MAKE) runner-verify runner-web-verify test-integration || status=$$?; \
	$(MAKE) test-env-down; \
	exit $$status

## runner-shell: [docker] Open an interactive shell in the runner.
runner-shell: test-env
	$(COMPOSE) exec runner bash
