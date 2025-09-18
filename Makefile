SHELL := /bin/bash

APP := job-queue-system
PKG := github.com/flyingrobots/go-redis-work-queue
VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)
GOFLAGS ?=

BIN_DIR := bin

.PHONY: all build test run lint tidy version clean

all: build

build: $(BIN_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP) ./cmd/$(APP)

.PHONY: build-tui tui-build
build-tui tui-build: $(BIN_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/tui ./cmd/tui

.PHONY: run-tui tui
run-tui tui: build-tui
	./bin/tui --config=config/config.yaml

run:
	./bin/$(APP) --role=all --config=config/config.yaml

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

lint:
	./scripts/check_yaml_newlines.py
	go run ./tools/requestidlint/cmd/requestidlint ./internal/admin-api

version:
	@echo $(VERSION)

.PHONY: clean
clean:
	rm -rf bin dist build out coverage *.coverprofile *.out .gocache

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

.PHONY: hooks
hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit
	@echo "Git hooks enabled (pre-commit updates progress bars and stages docs)."

.PHONY: mdlint
mdlint:
	@if ! command -v npx >/dev/null 2>&1; then \
		echo "npx not found. Please install Node.js to run markdownlint."; \
		exit 1; \
	fi
	@npx -y markdownlint-cli2 "**/*.md" "!**/node_modules/**"

.PHONY: mdlint-docs
mdlint-docs:
	@if ! command -v npx >/dev/null 2>&1; then \
		echo "npx not found. Please install Node.js to run markdownlint."; \
		exit 1; \
	fi
	@npx -y markdownlint-cli2 "docs/**/*.md"

.PHONY: mdlint-fix
mdlint-fix:
	@if ! command -v npx >/dev/null 2>&1; then \
		echo "npx not found. Please install Node.js to run markdownlint."; \
		exit 1; \
	fi
	@npx -y markdownlint-cli2 --fix "docs/**/*.md"

.PHONY: mdlint-docker
mdlint-docker:
	@docker run --rm -v "$(PWD)":/work -w /work node:20 \
	  npx -y markdownlint-cli2 "**/*.md" "!**/node_modules/**"
