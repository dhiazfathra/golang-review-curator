.PHONY: infra-up infra-down migrate migrate-down build test lint server-run worker-run

infra-up:
	docker compose -f deployments/docker-compose.yaml up -d

infra-down:
	docker compose -f deployments/docker-compose.yaml down

migrate:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

build:
	go build ./cmd/server/... ./cmd/worker/...

test:
	go test ./...

lint:
	golangci-lint run ./...

server-run:
	go run ./cmd/server/...

worker-run:
	go run ./cmd/worker/...
