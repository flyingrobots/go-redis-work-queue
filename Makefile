SHELL := /bin/bash

APP := job-queue-system
PKG := github.com/flyingrobots/go-redis-work-queue
VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)

.PHONY: all build test run lint tidy version

all: build

build:
	GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(APP) ./cmd/$(APP)

run:
	./bin/$(APP) --role=all --config=config/config.yaml

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

version:
	@echo $(VERSION)

.PHONY: hooks
hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit
	@echo "Git hooks enabled (pre-commit markdownlint autofix)."

.PHONY: mdlint
mdlint:
	@if ! command -v npx >/dev/null 2>&1; then \
		echo "npx not found. Please install Node.js to run markdownlint."; \
		exit 1; \
	fi
	@npx -y markdownlint-cli2 "**/*.md" "!**/node_modules/**"
