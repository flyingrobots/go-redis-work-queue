SHELL := /bin/bash

APP := job-queue-system
PKG := github.com/flyingrobots/go-redis-work-queue

.PHONY: all build test run lint tidy

all: build

build:
	GO111MODULE=on go build -o bin/$(APP) ./cmd/$(APP)

run:
	./bin/$(APP) --role=all --config=config/config.yaml

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

