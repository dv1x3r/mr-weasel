# syntax=docker/dockerfile:1

ARG GO_VERSION="1.23-alpine"

FROM golang:${GO_VERSION} AS build

RUN apk add --no-cache --update build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/mattn/go-sqlite3

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o ./bin/mr-weasel
RUN go test -v ./...

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=build /app/migrations /app/migrations
COPY --from=build /app/bin/mr-weasel /app/mr-weasel

CMD ["/app/mr-weasel"]
