.PHONY: tools/install tools/clean docs-generate setup update lint security build start run
TOOLS_BIN := $(shell pwd)/tools/bin

GOOS ?= $(shell go env GOOS)
EXT  :=

ifeq ($(GOOS),windows)
	EXT = .exe
endif

LINT_BIN := $(TOOLS_BIN)/golangci-lint$(EXT)
GOSEC_BIN := $(TOOLS_BIN)/gosec$(EXT)

GOLANGCI_LINT_VER := v2.1.6
GOSEC_VER         := v2.22.4

$(TOOLS_BIN):
	mkdir -p $(TOOLS_BIN)

tools/install: $(TOOLS_BIN) $(LINT_BIN) $(GOSEC_BIN)
	@echo "Instruments downloaded to $(TOOLS_BIN)"

$(LINT_BIN): | $(TOOLS_BIN)
	@echo "Installing golangci-lint@$(GOLANGCI_LINT_VER) into $(TOOLS_BIN)..."
	GOPROXY=direct GOBIN=$(TOOLS_BIN) \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VER)

$(GOSEC_BIN): | $(TOOLS_BIN)
	@echo "Installing gosec@$(GOSEC_VER) into $(TOOLS_BIN)..."
	GOPROXY=direct GOBIN=$(TOOLS_BIN) \
		go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VER)

tools/clean:
	rm -rf $(TOOLS_BIN)

lint: tools/install
ifeq ($(OS),Windows_NT)
	@echo "Running golangci-lint via go run on Windows..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VER) \
		run --config .golangci.yml ./... | tee .golangci-report.env
else
	@echo "Running golangci-lint from $(LINT_BIN)..."
	@$(LINT_BIN) run --config .golangci.yml ./... | tee .golangci-report.env
endif

security: tools/install
ifeq ($(OS),Windows_NT)
	@echo "Running gosec via go run on Windows..."
	@go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VER) ./...
else
	@echo "Running gosec from $(GOSEC_BIN)..."
	@$(GOSEC_BIN) ./...
endif

update:
	go mod tidy

build:
	go build -o pvz$(EXT) cmd/main.go

start:
	./pvz$(EXT)

setup:
	mkdir -p ../.git/hooks
	cp ../.hooks/pre-commit ../.git/hooks/pre-commit
	chmod +x ../.git/hooks/pre-commit

run: setup update lint security build start

.PHONY: docs-generate
docs-generate:
	@echo "generating per-package dev-docs"
	@mkdir -p docs/dev
	@for pkg in $$(go list ./...); do \
	  file="$$(echo $$pkg | sed 's#/#-#g').md"; \
	  echo "  - $$pkg → $$file"; \
	  go doc -all $$pkg > docs/dev/$$file; \
	done
	@echo "done: docs/dev/*.md"