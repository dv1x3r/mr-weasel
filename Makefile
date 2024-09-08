include .env

APP_NAME ?= main
GOOSE=goose -dir=${GOOSE_MIGRATION_DIR} ${GOOSE_DRIVER} ${GOOSE_DBSTRING}

build:
	go build -o bin/$(APP_NAME)

run:
	go build -o ./bin/$(APP_NAME) && ./bin/$(APP_NAME)

test:
	go test -v ./...

db-up:
	$(GOOSE) up

db-up-to:
	@read -p "Up to version: " VALUE; \
	$(GOOSE) up-to $$VALUE

db-up-by-one:
	$(GOOSE) up-by-one

db-down:
	$(GOOSE) down

db-down-to:
	@read -p "Down to version: " VALUE; \
	$(GOOSE) down-to $$VALUE

db-status:
	$(GOOSE) status 

db-reset:
	$(GOOSE) reset

db-create:
	@read -p "Migration name: " VALUE; \
	$(GOOSE) create "$$VALUE" sql

