NAME ?= $(error NAME not set in including Makefile)
LOWERNAME = $(shell echo $(NAME) | sed -e 's/-//g')
DIRNAME = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
.DEFAULT_GOAL = help

.PHONY: help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: db
db: ## Launches and migrates a development database
	COMPOSE_PROJECT_NAME=$(LOWERNAME) docker-compose stop
	COMPOSE_PROJECT_NAME=$(LOWERNAME) docker-compose rm --force
	COMPOSE_PROJECT_NAME=$(LOWERNAME) DOCKER_BUILDKIT=0 docker-compose build --no-cache
	COMPOSE_PROJECT_NAME=$(LOWERNAME) docker-compose up -d db

	until (docker exec $(LOWERNAME)-db-1 psql -q -U $$POSTGRES_USER -c "select now()" $$POSTGRES_DATABASE >/dev/null 2>&1); \
	do \
		echo 'Postgres not ready yet. Will try again in 1 second.'; sleep 1; \
	done; \

	docker exec $(LOWERNAME)-db-1 psql -q -U $$POSTGRES_USER -c "$$(cat ./migrations/000_init.sql)" $$POSTGRES_DATABASE
	COMPOSE_PROJECT_NAME=$(LOWERNAME) docker-compose up migrate

	for f in fixtures/*.sql; \
	do \
    echo "Running $$f..."; \
		docker exec $(LOWERNAME)-db-1 psql -q -U $$POSTGRES_USER $$POSTGRES_DATABASE -f /tmp/$$f; \
	done;
