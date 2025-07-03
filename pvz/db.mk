TOOLS_BIN   ?= ./tools/bin

GOOSE_BIN := $(TOOLS_BIN)/goose$(EXT)
GOOSE_VER := v3.16.0
GOOSE_DRIVER=postgres
DB_DSN ?= $(DB_WRITE_DSN)

TEST_DB_DSN = "postgres://test_user:test_pass@localhost:5455/pvz_test?sslmode=disable"

$(GOOSE_BIN):
	@mkdir -p $(dir $(GOOSE_BIN))
	@echo "Installing goose@$(GOOSE_VER) into $(TOOLS_BIN)"
	GOPROXY=direct GOBIN=$(TOOLS_BIN) \
		go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VER)


.PHONY: migrate-up
migrate-up: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations $(GOOSE_DRIVER) "$(DB_DSN)" up

.PHONY: migrate-status
migrate-status: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations $(GOOSE_DRIVER) "$(DB_DSN)" status

.PHONY: migrate-down
migrate-down: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations $(GOOSE_DRIVER) "$(DB_DSN)" down

.PHONY: migrate-new
migrate-new: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations create $(name) sql

.PHONY: test/migrate-up
test/migrate-up: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations $(GOOSE_DRIVER) "$(TEST_DB_DSN)" up

.PHONY: test/migrate-status
test/migrate-status: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations $(GOOSE_DRIVER) "$(TEST_DB_DSN)" status

.PHONY: test/migrate-down
test/migrate-down: $(GOOSE_BIN)
	@$(GOOSE_BIN) -dir migrations $(GOOSE_DRIVER) "$(TEST_DB_DSN)" down