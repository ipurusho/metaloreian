.PHONY: dev backend frontend db migrate clean build test test-backend test-frontend

dev: db backend frontend

db:
	docker compose up -d postgres

backend:
	set -a && [ -f .env ] && . ./.env; cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

migrate:
	docker compose exec postgres psql -U metaloreian -d metaloreian -f /docker-entrypoint-initdb.d/001_initial_schema.sql

clean:
	docker compose down -v

build:
	cd frontend && npm run build
	cd backend && CGO_ENABLED=0 go build -o metaloreian ./cmd/server

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

docker:
	docker compose up --build
