# syntax=docker/dockerfile:1

ARG GO_VERSION="1.23"

FROM golang:${GO_VERSION} AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/mr-weasel
RUN go test -v ./...

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/migrations /app/migrations
COPY --from=build /app/bin/mr-weasel /app/mr-weasel

CMD ["/app/mr-weasel"]
