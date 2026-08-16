GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: all api-check api-compat api-generate api-validate build check format format-check lint run test vet

all: check

## api-generate: Regenerate server and browser types from the OpenAPI source.
api-generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 --config api/openapi/oapi-codegen.yaml api/openapi/code-winch.yaml
	cd web && npm run api:generate

## api-validate: Parse and validate the OpenAPI document and contract fixtures.
api-validate:
	go test ./test/contract/openapi -run TestOpenAPIContract

## api-compat: Reject breaking changes relative to the committed v1 baseline.
api-compat:
	go run github.com/oasdiff/oasdiff@v1.11.7 breaking test/contract/openapi/v1.yaml api/openapi/code-winch.yaml --fail-on ERR

## api-check: Verify validation, compatibility, and deterministic generated output.
api-check: api-validate api-compat
	@tmp="$$(mktemp -d)"; trap 'rm -rf "$$tmp"' EXIT; \
	cp internal/adapters/transport/httpapi/types.gen.go "$$tmp/go"; \
	cp web/src/api/schema.ts "$$tmp/ts"; \
	$(MAKE) api-generate; \
	cmp "$$tmp/go" internal/adapters/transport/httpapi/types.gen.go; \
	cmp "$$tmp/ts" web/src/api/schema.ts

## format: Format all Go source files.
format:
	$(GO) fmt ./...

## format-check: Fail if any Go source file is not gofmt-formatted.
format-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "The following Go files are not formatted:" >&2; \
		echo "$$files" >&2; \
		echo "Run 'make format' to fix them." >&2; \
		exit 1; \
	fi

## vet: Report suspicious Go constructs.
vet:
	$(GO) vet ./...

## lint: Run the pinned repository-wide Go linter.
lint:
	$(GOLANGCI_LINT) run

## test: Run all Go unit tests.
test:
	$(GO) test ./...

## build: Build the daemon.
build:
	@output="$$(mktemp)"; \
	trap 'rm -f "$$output"' EXIT; \
	$(GO) build -o "$$output" ./cmd/winchd

## run: Start the daemon with configuration resolved from file and environment.
run:
	$(GO) run ./cmd/winchd

## check: Run every check required by CI.
check: api-check format-check vet lint test build
