PNPM ?= $(shell which pnpm 2>/dev/null || echo pnpm)

.PHONY: all setup ingest dev build test validate clean

all: build

setup:
	go mod download
	cd apps/web && $(PNPM) install --ignore-scripts

ingest:
	go run ./cmd/atlas ingest --source huggingface

dev:
	go run ./cmd/atlas serve --port 8080

build:
	cd apps/web && $(PNPM) build
	mkdir -p bin
	go build -o bin/atlas ./cmd/atlas

test:
	go test -v ./...
	cd apps/web && $(PNPM) test

validate:
	go run ./cmd/atlas validate

clean:
	rm -rf bin/
	rm -rf apps/web/dist/
	rm -f *.csv
