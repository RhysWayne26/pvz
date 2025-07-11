GOOS ?= $(shell go env GOOS)
EXT  :=

ifeq ($(GOOS),windows)
	EXT = .exe
endif


ifeq ($(GOOS),windows)
	EXT = .exe
endif

.PHONY: update build start run

update:
	go mod tidy

build:
	go build -o notifier$(EXT) cmd/main.go

start:
	./notifier$(EXT)

run: build start