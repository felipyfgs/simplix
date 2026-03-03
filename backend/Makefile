.PHONY: run build migrate-up migrate-down tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

migrate-create:
	@read -p "Migration name: " name; \
	go run ./cmd/migrate create $${name} sql
