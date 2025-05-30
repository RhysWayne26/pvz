.PHONY: docs docs-serve docs-generate setup update lint security build start run

GOOS ?= $(shell go env GOOS)
EXT =

ifeq ($(GOOS),windows)
EXT = .exe
endif

docs:
	@echo "Starting documentation server at http://localhost:6060"
	godoc -http=:6060

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
