.PHONY: run dev test generate migrate-up migrate-down bruno-env

ifneq (,$(wildcard .env))
include .env
export APP_ENV HTTP_ADDR DATABASE_URL JWT_SECRET JWT_TTL AUTO_MIGRATE
export OTEL_EXPORTER_OTLP_ENDPOINT OTEL_SERVICE_NAME
endif

run:
	go run ./cmd/api

dev: bruno-env
	@command -v air >/dev/null 2>&1 || { echo "Air is required: https://github.com/air-verse/air"; exit 1; }
	docker compose up -d postgres
	air -c .air.toml

test:
	go test ./...

generate:
	sqlc generate

migrate-up:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" down

bruno-env:
	@test -f bruno/environments/local.bru || cp bruno/environments/local.example.bru bruno/environments/local.bru
	@echo "Bruno local environment is ready: bruno/environments/local.bru"
