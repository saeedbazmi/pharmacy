.DEFAULT_GOAL := help
COMPOSE := docker compose

.PHONY: help up down logs ps migrate sqlc api worker web test lint fmt tidy seed-bulk create-ops-user seed-matches backup restore

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

up: ## Start db, api, worker and web
	$(COMPOSE) up -d --build

down: ## Stop everything (data volume is kept)
	$(COMPOSE) down

logs: ## Tail logs of all services
	$(COMPOSE) logs -f --tail=100

ps: ## Show service status
	$(COMPOSE) ps

migrate: ## Apply database migrations
	$(COMPOSE) run --rm migrate

sqlc: ## Regenerate type-safe query code from db/queries
	cd backend && sqlc generate

api: ## Run the api locally (needs DATABASE_URL)
	cd backend && go run ./cmd/api

worker: ## Run the worker locally (needs DATABASE_URL)
	cd backend && go run ./cmd/worker

web: ## Run the frontend dev server
	cd web && npm run dev

test: ## Run backend and frontend tests
	cd backend && go test ./...
	cd web && npm run test --if-present

lint: ## Vet backend and lint frontend
	cd backend && go vet ./...
	cd web && npm run lint

fmt: ## Format backend and frontend sources
	cd backend && go fmt ./...
	cd web && npm run format

tidy: ## Tidy Go module dependencies
	cd backend && go mod tidy

seed-bulk: ## Insert tens of thousands of realistic-ish products for search benchmarks
	$(COMPOSE) --profile seed run --rm seedbulk

create-ops-user: ## Create the first data-ops login (OPS_USERNAME / OPS_PASSWORD)
	$(COMPOSE) --profile ops run --rm opsuser

seed-matches: ## Queue 50 real source items into the match review queue
	$(COMPOSE) --profile ops run --rm seedmatches

backup: ## Dump Postgres into deploy/backups/
	bash deploy/backup.sh

restore: ## Restore BACKUP=path.sql.gz (CONFIRM=yes)
	bash deploy/restore.sh
