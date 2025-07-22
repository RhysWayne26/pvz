.DEFAULT_GOAL := help
COMPOSE_ENV_FILE := pvz/.env


.PHONY: help
help:
	@echo "Proxy into sub‑modules:"
	@echo "  make pvz-<target> # run <target> in pvz/"
	@echo "  make notifier-<target> # run <target> in notifier/"

.PHONY: pvz-% notifier-%
pvz-%:
	@$(MAKE) -C pvz $*

notifier-%:
	@$(MAKE) -C notifier $*

.PHONY: go-workspace
go-workspace:
	go work init ./pvz ./notifier

.PHONY: pvz-build notifier-build build-all
pvz-build:
	@$(MAKE) -C pvz build

notifier-build:
	@$(MAKE) -C notifier build

build-all: pvz-build notifier-build

.PHONY: docker-up docker-down docker-down-volumes docker-status

docker-up:
	docker-compose --env-file $(COMPOSE_ENV_FILE) up -d \
	  postgres-master postgres-slave \
	  kafka kafka-init kafka-ui \
	  pvz notifier migrator jaeger prometheus

docker-down:
	docker-compose --env-file $(COMPOSE_ENV_FILE) down \
	  postgres-master postgres-slave \
	  kafka kafka-init kafka-ui \
	  pvz notifier migrator jaeger prometheus

docker-down-volumes:
	docker-compose --env-file $(COMPOSE_ENV_FILE) down -v \
	  postgres-master postgres-slave \
	  kafka kafka-init kafka-ui \
	  pvz notifier migrator jaeger prometheus

docker-status:
	docker-compose --env-file $(COMPOSE_ENV_FILE) ps \
	  postgres-master postgres-slave \
	  kafka kafka-init kafka-ui \
	  pvz notifier migrator jaeger prometheus