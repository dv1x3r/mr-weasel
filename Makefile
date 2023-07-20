include .env

APP_NAME ?= main
GOOSE=goose -dir=./migrations ${GOOSE_DRIVER} ${GOOSE_DBSTRING}

build:
	go build -o bin/$(APP_NAME)

compile:
	GOOS=darwin GOARCH=arm64 go build -o ./bin/$(APP_NAME)-darwin-arm64
	GOOS=linux GOARCH=arm64 go build -o ./bin/$(APP_NAME)-linux-arm64
	GOOS=linux GOARCH=amd64 go build -o ./bin/$(APP_NAME)-linux-amd64
	GOOS=windows GOARCH=amd64 go build -o ./bin/$(APP_NAME)-windows-amd64.exe

run:
	go build -o ./bin/$(APP_NAME) && ./bin/$(APP_NAME)

db-up:
	$(GOOSE) up

db-up-by-one:
	$(GOOSE) up-by-one

db-up-to:
	@read -p "Up to version: " VALUE; \
	$(GOOSE) up-to $$VALUE

db-down:
	$(GOOSE) down

db-down-to:
	@read -p "Down to version: " VALUE; \
	$(GOOSE) down-to $$VALUE

db-reset:
	$(GOOSE) reset

db-status:
	$(GOOSE) status 

db-create:
	@read -p "Migration name: " VALUE; \
	$(GOOSE) create "$$VALUE" sql

.PHONY: build compile run
.PHONY: db-up db-up-by-one db-up-to db-down db-down-to
.PHONY: db-reset db-status db-create
 
