FROM golang:1.21 as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt update && \
    apt install -y curl && \
    curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh

COPY --from=build /app/server /app/server
COPY --from=build /app/migrations /app/migrations

CMD ["sh", "-c", "goose up && exec /app/server"]
