SHELL := /bin/bash
GOBIN := $(shell go env GOPATH)/bin
BINARY := tmp/server
DSN ?= postgres://postgres:postgres@localhost:5432/app?sslmode=disable
MIGRATIONS := backend/internal/repository/postgres/migrations

.PHONY: help tools db-up db-down dev build run sqlc migrate migrate-down test lint fmt docker clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

tools: ## Install Go dev tools (air, sqlc, goose, golangci-lint)
	go install github.com/air-verse/air@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

db-up: ## Start Postgres (docker compose)
	docker compose up -d

db-down: ## Stop Postgres
	docker compose down

dev: db-up ## Run backend (air, :3000) + frontend (bun, :4200) together
	@echo "Fiber on :3000 (air) + Angular on :4200 (bun, proxying /api)..."
	@trap 'kill 0' EXIT; \
		$(GOBIN)/air & \
		( cd frontend && bun run start ) & \
		wait

build: ## Build the single self-contained binary (frontend + backend)
	cd frontend && bun run build
	go build -ldflags="-s -w" -o $(BINARY) ./backend/cmd/server
	@echo "Built $(BINARY)"

run: build ## Build and run the single binary
	./$(BINARY)

sqlc: ## Regenerate sqlc code from queries + migrations
	$(GOBIN)/sqlc generate

migrate: ## Apply migrations manually (also auto-applied on startup)
	$(GOBIN)/goose -dir $(MIGRATIONS) postgres "$(DSN)" up

migrate-down: ## Roll back the last migration
	$(GOBIN)/goose -dir $(MIGRATIONS) postgres "$(DSN)" down

test: ## Run backend and frontend tests
	go test ./backend/...
	cd frontend && bunx ng test --watch=false

lint: ## Lint backend (golangci-lint) and frontend (eslint)
	$(GOBIN)/golangci-lint run ./backend/...
	cd frontend && bunx ng lint

fmt: ## Format Go and frontend code
	go fmt ./backend/...
	cd frontend && bunx prettier --write "src/**/*.{ts,html,scss}"

docker: ## Build the production Docker image
	docker build -t gofiber-angular-spa .

clean: ## Remove build artifacts
	rm -rf tmp
	find backend/web/dist/browser -type f ! -name index.html -delete 2>/dev/null || true
