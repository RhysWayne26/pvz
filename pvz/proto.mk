PROTO_DIR := api/grpc
PROTO_FILES := orders.proto admin.proto
THIRD_PARTY_DIR := third_party
SWAGGER_OUT := docs/swagger
PROTO_OUT := internal/gen/orders
ADMIN_PROTO_OUT := internal/gen/admin

TOOLS_BIN := tools/bin

PROTOC_GEN_GO_VER := v1.36.5
PROTOC_GEN_GO_GRPC_VER := v1.5.1
GRPC_GATEWAY_VER := v2.26.3
OPENAPI_VER := v2.26.3
PROTOC_GEN_VALIDATE := v1.0.2

PROTOC_VERSION := 25.3
PROTOC_DIR := tools/protoc
PROTOC_BIN := $(PROTOC_DIR)/bin/protoc

UNAME_S := $(shell uname -s)
UNAME_P := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
	PROTOC_OS := linux
endif
ifeq ($(UNAME_S),Darwin)
	PROTOC_OS := osx
endif
ifeq ($(OS),Windows_NT)
	PROTOC_OS := win
	PROTOC_BIN := $(PROTOC_DIR)/bin/protoc.exe
endif

ifeq ($(UNAME_P),x86_64)
	PROTOC_ARCH := x86_64
else
	PROTOC_ARCH := $(UNAME_P)
endif

PROTOC_ZIP := protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip
PROTOC_URL := https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP)

.PHONY: all
all: proto-generate proto-generate-admin

.PHONY: $(TOOLS_BIN)
$(TOOLS_BIN):
	mkdir -p $(TOOLS_BIN)

.PHONY: tools-protoc
tools-protoc:
	@echo "Downloading protoc $(PROTOC_VERSION) for $(PROTOC_OS)-$(PROTOC_ARCH)..."
	@rm -rf $(PROTOC_DIR)
	@mkdir -p $(PROTOC_DIR)
	@curl -sSL $(PROTOC_URL) -o $(PROTOC_DIR)/$(PROTOC_ZIP)
ifeq ($(PROTOC_OS),win)
	@powershell -Command "Remove-Item -Path '$(PROTOC_DIR)/*' -Recurse -Force -ErrorAction SilentlyContinue; Expand-Archive -Path '$(PROTOC_DIR)/$(PROTOC_ZIP)' -DestinationPath '$(PROTOC_DIR)' -Force"
else
	@unzip -o $(PROTOC_DIR)/$(PROTOC_ZIP) -d $(PROTOC_DIR)
endif
	@rm -f $(PROTOC_DIR)/$(PROTOC_ZIP)
	@echo "protoc downloaded to $(PROTOC_BIN)"

.PHONY: tools-proto
tools-proto: $(TOOLS_BIN)
	@echo "Installing protobuf plugins..."
	GOBIN=$(TOOLS_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VER)
	GOBIN=$(TOOLS_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VER)
	GOBIN=$(TOOLS_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GRPC_GATEWAY_VER)
	GOBIN=$(TOOLS_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(OPENAPI_VER)
	GOBIN=$(TOOLS_BIN) go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE)
	@echo "Finished installing protobuf plugins."

.PHONY: proto-setup-third-party
proto-setup-third-party:
	@echo "Downloading Google API .proto files..."
	@mkdir -p $(THIRD_PARTY_DIR)/google/api
	@curl -sL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto \
		-o $(THIRD_PARTY_DIR)/google/api/annotations.proto
	@curl -sL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto \
		-o $(THIRD_PARTY_DIR)/google/api/http.proto
	@echo "Downloading protoc-gen-validate .proto file..."
	@mkdir -p $(THIRD_PARTY_DIR)/validate
	@curl -sL https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/main/validate/validate.proto \
		-o $(THIRD_PARTY_DIR)/validate/validate.proto
	@echo "Finished setting up third-party proto files."

.PHONY: $(PROTO_OUT)
$(PROTO_OUT):
	@mkdir -p $@

.PHONY: $(ADMIN_PROTO_OUT)
$(ADMIN_PROTO_OUT):
	@mkdir -p $@


.PHONY: $(SWAGGER_OUT)
$(SWAGGER_OUT):
	@mkdir -p $@

.PHONY: proto-generate
proto-generate: tools-proto tools-protoc proto-setup-third-party $(PROTO_OUT) $(SWAGGER_OUT)
	@echo "Generating code from .proto..."
	cd $(PROTO_DIR) && \
	PATH=../../$(PROTOC_DIR)/bin:$$PATH \
	../../$(PROTOC_BIN) \
	  -I. \
	  -I../../$(THIRD_PARTY_DIR) \
	  -I../../$(THIRD_PARTY_DIR)/validate \
	  -I../../$(PROTOC_DIR)/include \
      --go_out=../../$(PROTO_OUT) --go_opt=paths=source_relative \
      --go-grpc_out=../../$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
      --grpc-gateway_out=../../$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
	  --openapiv2_out=../../$(SWAGGER_OUT) \
      --openapiv2_opt logtostderr=true,json_names_for_fields=false \
	  --validate_out="lang=go,paths=source_relative:../../$(PROTO_OUT)" \
	  --plugin=protoc-gen-validate=$(abspath $(TOOLS_BIN))/protoc-gen-validate \
	  --plugin=protoc-gen-go=$(abspath $(TOOLS_BIN))/protoc-gen-go \
	  --plugin=protoc-gen-go-grpc=$(abspath $(TOOLS_BIN))/protoc-gen-go-grpc \
	  --plugin=protoc-gen-grpc-gateway=$(abspath $(TOOLS_BIN))/protoc-gen-grpc-gateway \
	  --plugin=protoc-gen-openapiv2=$(abspath $(TOOLS_BIN))/protoc-gen-openapiv2 \
	  $(PROTO_FILES)
	@echo "Code generated in $(PROTO_OUT)/ and swagger in $(SWAGGER_OUT)/."

.PHONY: proto-generate-admin
proto-generate-admin: tools-proto tools-protoc proto-setup-third-party $(ADMIN_PROTO_OUT)
	@echo "Generating admin proto code..."
	cd $(PROTO_DIR) && \
	PATH=../../$(PROTOC_DIR)/bin:$$PATH \
	../../$(PROTOC_BIN) \
		-I. \
		-I../../$(THIRD_PARTY_DIR) \
		-I../../$(PROTOC_DIR)/include \
		--go_out=../../$(ADMIN_PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=../../$(ADMIN_PROTO_OUT) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=../../$(ADMIN_PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=../../$(SWAGGER_OUT) \
		--openapiv2_opt logtostderr=true,json_names_for_fields=false \
		--plugin=protoc-gen-go=$(abspath $(TOOLS_BIN))/protoc-gen-go \
		--plugin=protoc-gen-go-grpc=$(abspath $(TOOLS_BIN))/protoc-gen-go-grpc \
		--plugin=protoc-gen-grpc-gateway=$(abspath $(TOOLS_BIN))/protoc-gen-grpc-gateway \
		--plugin=protoc-gen-openapiv2=$(abspath $(TOOLS_BIN))/protoc-gen-openapiv2 \
		admin.proto
	@echo "Admin proto code generated in $(ADMIN_PROTO_OUT)/"

.PHONY: proto-update-deps
proto-update-deps: tools-proto
	@echo "Proto dependencies are up to date."

.PHONY: proto-clean
proto-clean:
	@echo "Cleaning generated files..."
	rm -rf $(PROTO_OUT)/*
	rm -rf $(ADMIN_PROTO_OUT)/*
	rm -rf $(SWAGGER_OUT)/*
	@echo "Finished cleaning."

.PHONY: proto-check
proto-check:
	@echo "Checking proto dependencies..."
	@test -f $(PROTOC_BIN) || { echo "protoc is not installed; run make tools/protoc"; exit 1; }
	@test -f $(TOOLS_BIN)/protoc-gen-go || { echo "protoc-gen-go not found; run make tools/proto"; exit 1; }
	@test -f $(TOOLS_BIN)/protoc-gen-go-grpc || { echo "protoc-gen-go-grpc not found; run make tools/proto"; exit 1; }
	@test -f $(TOOLS_BIN)/protoc-gen-grpc-gateway || { echo "protoc-gen-grpc-gateway not found; run make tools/proto"; exit 1; }
	@test -f $(TOOLS_BIN)/protoc-gen-openapiv2 || { echo "protoc-gen-openapiv2 not found; run make tools/proto"; exit 1; }
	@test -f $(TOOLS_BIN)/protoc-gen-validate || { echo "protoc-gen-validate not found; run make tools/proto"; exit 1; }
	@echo "All proto dependencies are present."