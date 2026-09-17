.DEFAULT_GOAL := help
COMPOSE := docker compose

.PHONY: help up down logs ps migrate sqlc api worker web test lint fmt tidy

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
