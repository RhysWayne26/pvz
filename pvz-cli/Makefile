.PHONY: docs-generate setup update lint security build start run

GOOS ?= $(shell go env GOOS)
EXT =

ifeq ($(GOOS),windows)
EXT = .exe
endif

setup:
	mkdir -p ../.git/hooks
	cp ../.hooks/pre-commit ../.git/hooks/pre-commit
	chmod +x ../.git/hooks/pre-commit

update:
	go mod tidy

lint:
	golangci-lint run --config .golangci.yml ./... | tee .golangci-report.env

security:
	gosec ./...

build:
	go build -o pvz$(EXT) cmd/main.go

start:
	./pvz$(EXT)

run: setup update lint security build start

docs-generate:
	@echo "Generating per-package dev-docs…"
	@mkdir -p docs/dev
	@for pkg in $$(go list ./...); do
	  file="$$(echo $$pkg | sed 's#/#-#g').md";
	  echo "  - $$pkg → $$file";
	  go doc -all $$pkg > docs/dev/$$file;
	done
	@echo "Done: docs/dev/*.md"