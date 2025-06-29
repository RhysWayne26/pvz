MINIMOCK_VER := v3.4.5
MINIMOCK := go run github.com/gojuno/minimock/v3/cmd/minimock@$(MINIMOCK_VER)
USECASES_MOCK_OUTPUT_DIR := internal/usecases/mocks
MODULE := $(shell go list -m)
ALLURE_CLI_VER := 2.25.0
ALLURE_RESULTS := $(shell pwd)/allure-results
ALLURE_REPORT := $(shell pwd)/allure-report
ALLURE_DIR := tools/allure
ALLURE_BIN := $(ALLURE_DIR)/allure-$(ALLURE_CLI_VER)/bin/allure
ALLURE_GO_VER := v0.6.34

define gen_mock
	@$(MINIMOCK) -i $(MODULE)/$1 -o $(MOCK_OUTPUT)/$2
endef

.PHONY: mocks mocks/generate
mocks/generate:
	@echo "generating minimock interfaces into $(USECASES_MOCK_OUTPUT_DIR)"
	mkdir -p $(USECASES_MOCK_OUTPUT_DIR)
	$(eval MOCK_OUTPUT := $(USECASES_MOCK_OUTPUT_DIR))

	# Repositories
	$(call gen_mock,internal/data/repositories.HistoryRepository,history_repository_mock.go)
	$(call gen_mock,internal/data/repositories.OrderRepository,order_repository_mock.go)

	# Validators
	$(call gen_mock,internal/usecases/services/validators.OrderValidator,order_validator_mock.go)
	$(call gen_mock,internal/usecases/services/validators.PackageValidator,package_validator_mock.go)

	# Strategies
	$(call gen_mock,internal/usecases/services/strategies.PricingStrategy,pricing_strategy_mock.go)

	# Services
	$(call gen_mock,internal/usecases/services.OrderService,order_service_mock.go)
	$(call gen_mock,internal/usecases/services.HistoryService,history_service_mock.go)
	$(call gen_mock,internal/usecases/services.PackagePricingService,package_pricing_service_mock.go)

mocks: mocks/generate

.PHONY: mocks/clean
mocks/clean:
	@echo "cleaning mocks from $(USECASES_MOCK_OUTPUT_DIR)"
	rm -f $(USECASES_MOCK_OUTPUT_DIR)/*_mock.go

COVER_PKGS := $(shell \
	go list ./internal/usecases/... \
	| grep -v /mocks \
	| grep -v /requests \
	| grep -v /responses \
	| grep -v /strategies \
	| grep -v /builders \
)

.PHONY: cover
cover:
	@echo "running coverage for:"
	@printf "   %s\n" $(COVER_PKGS)
	@go test -coverprofile=coverage.out -covermode=atomic $(COVER_PKGS)
	@go tool cover -func=coverage.out

.PHONY: cover/html
cover/html: cover
	@go tool cover -html=coverage.out -o coverage.html
	@echo "open coverage.html to view the report"

.PHONY: e2e/test
e2e/test:
	@echo "running e2e tests with Allure output"
	@mkdir -p $(ALLURE_RESULTS)
	ALLURE_OUTPUT_PATH=$(shell pwd) go test -v ./tests/e2e

.PHONY: int/test
int/test:
	@echo "running integration tests with Allure output"
	@mkdir -p $(ALLURE_RESULTS)
	ALLURE_OUTPUT_PATH=$(shell pwd) go test -v ./tests/integration/...


.PHONY: allure/install
allure/install:
	@echo "Installing Allure tools..."
	@echo "allure-go library should be in go.mod dependencies"
	@if [ ! -f $(ALLURE_BIN) ]; then \
		echo "Installing Allure CLI..."; \
		mkdir -p $(ALLURE_DIR); \
		echo "Downloading Allure CLI $(ALLURE_CLI_VER)..."; \
		curl -L "https://github.com/allure-framework/allure2/releases/download/$(ALLURE_CLI_VER)/allure-$(ALLURE_CLI_VER).tgz" | tar -xz -C $(ALLURE_DIR); \
		echo "Allure CLI installed at $(ALLURE_BIN)"; \
	else \
		echo "Allure CLI already installed at $(ALLURE_BIN)"; \
	fi
	@echo "Allure tools ready"

.PHONY: allure/report
allure/report: allure/install
	@echo "Generating Allure report..."
	@mkdir -p $(ALLURE_REPORT)
	@$(ALLURE_BIN) generate $(ALLURE_RESULTS) -o $(ALLURE_REPORT) --clean

.PHONY: allure/open
allure/open: allure/report
	@echo "Opening Allure report..."
	@$(ALLURE_BIN) open $(ALLURE_REPORT) || \
		echo "Report generated at: $(ALLURE_REPORT)/index.html"

.PHONY: test/all
test/all: allure/install e2e/test int/test allure/report allure/open