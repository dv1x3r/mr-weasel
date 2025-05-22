# syntax=docker/dockerfile:1

ARG GO_VERSION="1.24-alpine"

FROM golang:${GO_VERSION} AS build

WORKDIR /app

RUN apk add --no-cache --update build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/app ./cmd/app

RUN go test -v ./...

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=build /app/build/app /app/mr-weasel

CMD ["/app/mr-weasel"]
