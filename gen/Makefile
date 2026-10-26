# Makefile for Clever Cloud SDK

# The document the SDK is generated from: AXO's unified OpenAPI specification, which
# merges every Clever Cloud API source (the Rust port's own document, the legacy cc-api
# and OVD documents, the addon providers' routes). The SDK used to be generated from the
# OVD document alone. Override on the command line to generate from another checkout:
#   make build OPENAPI=/path/to/unified-spec.json
# (SPEC is the variables file of `make generate-app`, a different document.)
OPENAPI ?= ../../../rust/axo/crates/api/api/openapi-reference/unified-spec.json

.PHONY: help
help:
	@echo "Clever Cloud SDK - Code Generation"
	@echo ""
	@echo "OpenAPI document: $(OPENAPI)"
	@echo ""
	@echo "Available targets:"
	@echo "  make all                - Generate all code and run checks"
	@echo "  make build              - Generate models, services and builder"
	@echo "  make generate-models    - Generate models from OpenAPI spec"
	@echo "  make generate-services  - Generate services from OpenAPI spec"
	@echo "  make generate-builder   - Generate builder from services"
	@echo "  make generate-app SPEC=<file> - Generate app environment variables"
	@echo "  make check              - Check that generators and SDK compile"
	@echo "  make check-generators   - Check that all generators compile"
	@echo "  make check-sdk          - Check that SDK compiles"
	@echo "  make test-generators    - Run the generators' own tests"
	@echo "  make test-sdk           - Run SDK tests"
	@echo "  make clean              - Remove all generated files"
	@echo ""

.PHONY: all
all: build check

# `clean` first so files for schemas/operations removed from the spec don't
# linger and clash with current output (e.g. a deleted union leaving a stale
# *_union.go that redeclares an identifier). Generators only ever write files
# for schemas present in the spec — they never delete — so orphans must be
# swept before each build.
.PHONY: build
build: clean generate-models generate-services generate-builder

.PHONY: generate-models
generate-models:
	@echo "🔨 Generating models..."
	go run \
		./cmd/generate-models/main.go \
		--spec $(OPENAPI) \
		--output ../models

.PHONY: generate-services
generate-services:
	@echo "🔨 Generating services..."
	go run \
		./cmd/generate-services/main.go \
		--spec $(OPENAPI) \
		--output ../services

.PHONY: generate-builder
generate-builder:
	@echo "🔨 Generating builder..."
	go run \
		./cmd/generate-builder/main.go \
		--spec $(OPENAPI) \
		--output ../builder.go

# The description of the application environment variables is not part of
# this repository: pass its path with SPEC=<file>.
.PHONY: generate-app
generate-app:
	@test -n "$(SPEC)" || { echo "SPEC=<file> is required: path to the JSON description of the application environment variables"; exit 1; }
	@echo "🔨 Generating app environment variables..."
	go run \
		./cmd/generate-app/main.go \
		--spec $(SPEC) \
		--output ../app/variables.go

.PHONY: check
check: check-generators check-sdk

.PHONY: check-generators
check-generators:
	@echo "✅ Checking generators compile..."
	@go build -o /dev/null ./cmd/generate-models
	@go build -o /dev/null ./cmd/generate-services
	@go build -o /dev/null ./cmd/generate-builder
	@go build -o /dev/null ./cmd/generate-app
	@echo "✅ All generators compile successfully"

.PHONY: check-sdk
check-sdk:
	@echo "✅ Checking SDK compiles..."
	@cd .. && go build ./...
	@echo "✅ SDK compiles successfully"

.PHONY: test-generators
test-generators:
	@echo "🧪 Running generator tests..."
	@go test ./internal/...

.PHONY: test-sdk
test-sdk:
	@echo "🧪 Running SDK tests..."
	@cd .. && go test ./...

.PHONY: clean
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -f ../builder.go
	@# Drop generated *.go files but keep hand-written tests/examples/fixtures.
	@# Anything regen-able (json_helpers.go, *_union.go, *_alias.go, model structs)
	@# is rewritten on the next `make generate-models`.
	@find ../models -type f -name '*.go' ! -name '*_test.go' -delete 2>/dev/null || true
	@find ../services -type f -name '*.go' ! -name '*_test.go' -delete 2>/dev/null || true
	@# A package the spec no longer tags leaves an empty directory behind; sweep those too.
	@find ../services -mindepth 1 -type d -empty -delete 2>/dev/null || true
	@echo "✅ Clean complete"
