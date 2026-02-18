APP_NAME=simorq_backend
COMPOSE_FILE=docker-compose.yml
DB_CONTAINER=$(APP_NAME)-postgres
REDIS_CONTAINER=$(APP_NAME)-redis
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=simorq_db
DB_PORT=5432
REDIS_PORT=6379

# pass config like `make run CONFIG=./config.yaml`
CONFIG ?= config.yaml

# --- Colors for help output ---
ESC    := $(shell printf '\033')
RESET  := $(ESC)[0m
CYAN   := $(ESC)[36m
YELLOW := $(ESC)[33m
GREEN  := $(ESC)[32m
BLUE   := $(ESC)[34m
RED    := $(ESC)[31m

# --- .PHONY declarations ---
.PHONY: build run http-start test clean install tidy fmt vet lint entgen docgen help
.PHONY: db-start db-stop db-restart db-logs db-shell db-status db-clean db-init db-migrate
.PHONY: redis-start redis-stop redis-restart redis-logs redis-shell redis-status redis-clean
.PHONY: dev down logs docker-build rebuild

# --- Application Targets ----------------------------------------

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o $(APP_NAME) .
	@echo "Build complete!"

run: ## Run the fiber backend (http start)
	@echo "Starting $(APP_NAME)..."
	@go run . --config "$(CONFIG)" http start

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Remove built binary
	@echo "Cleaning..."
	@rm -f "$(APP_NAME)"
	@echo "Clean complete!"

tidy: ## go mod tidy
	@echo "Tidying modules..."
	@go mod tidy

fmt: ## go fmt ./...
	@echo "Formatting..."
	@go fmt ./...

vet: ## go vet ./...
	@echo "Vetting..."
	@go vet ./...

lint: ## vet + test
	@echo "Linting (vet + test)..."
	@go vet ./...
	@go test ./...

# --- Codegen ------------------------------------------------

entgen: ## Generate Ent code
	@echo "Generating Ent code..."
	@go generate ./internal/repo/...
	@echo "Ent code generation complete!"

docgen: ## Generate Cobra CLI docs
	@echo "Generating Cobra CLI docs..."
	@go run . --config "$(CONFIG)" system gendocs
	@echo "CLI documentation complete!"

migrate: ## Run database migrations
	@echo "Running DB migrations..."
	@go run . --config "$(CONFIG)" system migrate
	@echo "Migrations complete!"

# --- Database / Postgres Targets --------------------------

db-start: ## Start PostgreSQL 18 development database
	@echo "$(BLUE)Starting PostgreSQL 18 database ($(DB_CONTAINER))...$(RESET)"
	@docker rm -f $(DB_CONTAINER) 2>/dev/null || true
	@docker run -d \
		--name $(DB_CONTAINER) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-v $(DB_CONTAINER)-data:/var/lib/postgresql \
		postgres:18
	@echo "$(GREEN)PostgreSQL started!$(RESET)"
	@echo "Connection: postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)"

db-stop: ## Stop the development database
	@echo "$(YELLOW)Stopping PostgreSQL database...$(RESET)"
	@docker stop $(DB_CONTAINER) 2>/dev/null || true
	@docker rm $(DB_CONTAINER) 2>/dev/null || true
	@echo "$(GREEN)Database stopped!$(RESET)"

db-restart: db-stop db-start ## Restart the development database

db-logs: ## Show database logs
	@echo "$(BLUE)PostgreSQL logs:$(RESET)"
	@docker logs -f $(DB_CONTAINER)

db-shell: ## Open psql shell in the database
	@echo "$(BLUE)Connecting to PostgreSQL...$(RESET)"
	@docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

db-status: ## Check database status
	@echo "$(BLUE)Checking PostgreSQL status...$(RESET)"
	@docker ps -f name=$(DB_CONTAINER) --format "table {{.Names}}\t{{.Status}}" || echo "$(RED)Database container not running$(RESET)"

db-clean: ## Remove database container and data volume
	@echo "$(RED)Removing database container and data...$(RESET)"
	@docker stop $(DB_CONTAINER) 2>/dev/null || true
	@docker rm $(DB_CONTAINER) 2>/dev/null || true
	@docker volume rm $(DB_CONTAINER)-data 2>/dev/null || true
	@echo "$(GREEN)Database cleaned!$(RESET)"

db-init: ## Initialize all three application databases (main, context, casbin)
	@echo "$(BLUE)Waiting for PostgreSQL to be ready...$(RESET)"
	@sleep 3
	@echo "$(BLUE)Initializing databases...$(RESET)"
	@go run . system init
	@echo "$(GREEN)Databases initialized!$(RESET)"

db-init-migrate: db-init migrate ## Initialize databases and run migrations

# --- Redis Targets ------------------------------------------

redis-start: ## Start Redis 7 development cache
	@echo "$(BLUE)Starting Redis 7 ($(REDIS_CONTAINER))...$(RESET)"
	@docker rm -f $(REDIS_CONTAINER) 2>/dev/null || true
	@docker run -d \
		--name $(REDIS_CONTAINER) \
		-p $(REDIS_PORT):6379 \
		-v $(REDIS_CONTAINER)-data:/data \
		redis:7-alpine redis-server --appendonly yes
	@echo "$(GREEN)Redis started!$(RESET)"
	@echo "Connection: redis://localhost:$(REDIS_PORT)"

redis-stop: ## Stop the development Redis
	@echo "$(YELLOW)Stopping Redis...$(RESET)"
	@docker stop $(REDIS_CONTAINER) 2>/dev/null || true
	@docker rm $(REDIS_CONTAINER) 2>/dev/null || true
	@echo "$(GREEN)Redis stopped!$(RESET)"

redis-restart: redis-stop redis-start ## Restart the development Redis

redis-logs: ## Show Redis logs
	@echo "$(BLUE)Redis logs:$(RESET)"
	@docker logs -f $(REDIS_CONTAINER)

redis-shell: ## Open redis-cli shell
	@echo "$(BLUE)Connecting to Redis...$(RESET)"
	@docker exec -it $(REDIS_CONTAINER) redis-cli

redis-status: ## Check Redis status
	@echo "$(BLUE)Checking Redis status...$(RESET)"
	@docker ps -f name=$(REDIS_CONTAINER) --format "table {{.Names}}\t{{.Status}}" || echo "$(RED)Redis container not running$(RESET)"

redis-clean: ## Remove Redis container and data volume
	@echo "$(RED)Removing Redis container and data...$(RESET)"
	@docker stop $(REDIS_CONTAINER) 2>/dev/null || true
	@docker rm $(REDIS_CONTAINER) 2>/dev/null || true
	@docker volume rm $(REDIS_CONTAINER)-data 2>/dev/null || true
	@echo "$(GREEN)Redis cleaned!$(RESET)"

# --- Docker Compose Targets --------------------

dev: ## Start full dev environment (postgres + redis + app with Air hot reload)
	@echo "$(BLUE)Starting development environment...$(RESET)"
	@docker compose -f $(COMPOSE_FILE) up --build
	@echo "$(GREEN)Development environment started!$(RESET)"
	@echo "$(CYAN)Service URLs:$(RESET)"
	@echo "  HTTP Server: http://localhost:8080"
	@echo "  PostgreSQL:  postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)"
	@echo "  Redis:       redis://localhost:$(REDIS_PORT)"

down: ## Stop the Docker Compose environment
	@echo "$(YELLOW)Stopping Docker Compose environment...$(RESET)"
	@docker compose -f $(COMPOSE_FILE) down
	@echo "$(GREEN)Environment stopped!$(RESET)"

logs: ## Tail logs from the app container
	@docker compose -f $(COMPOSE_FILE) logs -f app

docker-build: ## Build Docker images for the application
	@echo "$(BLUE)Building Docker images...$(RESET)"
	@docker compose -f $(COMPOSE_FILE) build
	@echo "$(GREEN)Docker images built!$(RESET)"

rebuild: down docker-build dev ## Tear down, rebuild images, and start dev environment

restart: down dev ## Restart the full dev environment

# --- Help ---------------------------------------------------

help: ## Show this help message
	@printf '$(CYAN)Simorq Backend$(RESET)\n\n'
	@printf 'Usage:\n'
	@printf '  $(YELLOW)make$(RESET) $(GREEN)<target>$(RESET)\n\n'
	@printf 'Targets:\n'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Default target
.DEFAULT_GOAL := help