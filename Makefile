SHELL := /usr/bin/bash

PROJECT_ROOT := $(CURDIR)

.PHONY: all build test tidy \
	build-orders build-dispatcher \
	run-orders run-dispatcher

all: build

## Dependency housekeeping
tidy:
	cd $(PROJECT_ROOT) && go mod tidy

## Build binaries
build: build-orders build-dispatcher

build-orders:
	cd $(PROJECT_ROOT) && go build -o bin/orders ./orders

build-dispatcher:
	cd $(PROJECT_ROOT) && go build -o bin/dispatcher ./dispatcher

## Tests
test:
	cd $(PROJECT_ROOT) && go test ./...

## Run services (use env vars for config)
run-orders:
	cd $(PROJECT_ROOT) && go run ./orders

run-dispatcher:
	cd $(PROJECT_ROOT) && go run ./dispatcher



