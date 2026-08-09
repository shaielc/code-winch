GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: all build check format format-check lint test vet

all: check

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

## check: Run every check required by CI.
check: format-check vet lint test build
