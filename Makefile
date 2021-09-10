export UID=$(shell id -u)
export GID=$(shell id -g)

VERSION := latest

build:
	docker build --build-arg APP_NAME=$(APP_NAME) -t $(APP_NAME):$(VERSION) .

DOCKER_COMPOSE=UID=$(UID) GID=$(GID) docker-compose -p $(APP_NAME)
up:
	@if [ "$(APP_ENV)" = "development" ] || [ "$(APP_ENV)" = "staging" ] || [ "$(APP_ENV)" = "production" ]; then \
		$(DOCKER_COMPOSE) -f docker-compose.yaml up -d; \
	else \
		$(DOCKER_COMPOSE) up -d; \
	fi

down:
	@if [ "$(APP_ENV)" = "development" ] || [ "$(APP_ENV)" = "staging" ] || [ "$(APP_ENV)" = "production" ]; then \
		$(DOCKER_COMPOSE) -f docker-compose.yaml down -v --remove-orphans; \
	else \
		$(DOCKER_COMPOSE) down -v --remove-orphans; \
	fi

clean:
	docker rmi -f $(APP_NAME)-dev:latest
	docker rmi -f $(APP_NAME):latest
