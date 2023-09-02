include .env

APP_NAME ?= main
GOOSE=goose -dir=./migrations ${GOOSE_DRIVER} ${GOOSE_DBSTRING}

build:
	go build -o bin/$(APP_NAME)

run:
	go build -o ./bin/$(APP_NAME) && ./bin/$(APP_NAME)

test:
	go test -v ./...

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

go-tools:
	go install golang.org/x/tools/gopls@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: build compile run test
.PHONY: db-up db-up-by-one db-up-to db-down db-down-to
.PHONY: db-reset db-status db-create
.PHONY: go-install
 
