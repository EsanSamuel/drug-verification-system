.PHONY: all build run test clean docker-up docker-down sqlc-gen tidy migrate migrate-down migrate-status

# Binary name
BIN_DIR=bin
BINARY_NAME=$(BIN_DIR)/server

all: tidy test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY_NAME) ./cmd/server

run:
	go run ./cmd/server

migrate:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

test:
	go test -v -race ./...

tidy:
	go mod tidy

sqlc-gen:
	sqlc generate

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

clean:
	rm -rf $(BIN_DIR)

