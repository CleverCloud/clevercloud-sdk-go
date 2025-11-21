# Makefile for Clever Cloud SDK

.PHONY: help
help:
	@echo "Clever Cloud SDK - Code Generation"
	@echo ""
	@echo "Available targets:"
	@echo "  make all                - Generate all code and run checks"
	@echo "  make build              - Generate models, services and builder"
	@echo "  make generate-models    - Generate models from OpenAPI spec"
	@echo "  make generate-services  - Generate services from OpenAPI spec"
	@echo "  make generate-builder   - Generate builder from services"
	@echo "  make check              - Check that generators and SDK compile"
	@echo "  make check-generators   - Check that all generators compile"
	@echo "  make check-sdk          - Check that SDK compiles"
	@echo "  make test-sdk           - Run SDK tests"
	@echo "  make clean              - Remove all generated files"
	@echo ""

.PHONY: all
all: build check

.PHONY: build
build: generate-models generate-services generate-builder

.PHONY: generate-models
generate-models:
	@echo "🔨 Generating models..."
	go run \
		./cmd/generate-models/main.go \
		--spec ../../../scala/ovd3/server/openapi.all.json \
		--output ../models

.PHONY: generate-services
generate-services:
	@echo "🔨 Generating services..."
	go run \
		./cmd/generate-services/main.go \
		--spec ../../../scala/ovd3/server/openapi.all.json \
		--output ../services

.PHONY: generate-builder
generate-builder:
	@echo "🔨 Generating builder..."
	go run \
		./cmd/generate-builder/main.go \
		--spec ../../../scala/ovd3/server/openapi.all.json \
		--output ../builder.go

.PHONY: check
check: check-generators check-sdk

.PHONY: check-generators
check-generators:
	@echo "✅ Checking generators compile..."
	@go build -o /dev/null ./cmd/generate-models
	@go build -o /dev/null ./cmd/generate-services
	@go build -o /dev/null ./cmd/generate-builder
	@echo "✅ All generators compile successfully"

.PHONY: check-sdk
check-sdk:
	@echo "✅ Checking SDK compiles..."
	@cd .. && go build ./...
	@echo "✅ SDK compiles successfully"

.PHONY: test-sdk
test-sdk:
	@echo "🧪 Running SDK tests..."
	@cd .. && go test ./...

.PHONY: clean
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -f ../builder.go
	@rm -rf ../models
	@rm -rf ../services
	@echo "✅ Clean complete"
