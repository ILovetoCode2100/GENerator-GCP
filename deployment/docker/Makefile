# Virtuoso Docker Makefile
.PHONY: help build up down logs shell clean prod-build prod-up prod-down test

# Default environment
ENV_FILE ?= .env
COMPOSE_FILE ?= ../../docker-compose.yml
PROD_COMPOSE_FILE ?= docker-compose.prod.yml

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
build: ## Build all services for development
	@echo "${YELLOW}Building development services...${NC}"
	docker-compose -f $(COMPOSE_FILE) build

up: ## Start all services in development mode
	@echo "${GREEN}Starting development services...${NC}"
	docker-compose -f $(COMPOSE_FILE) up -d
	@echo "${GREEN}Services started! API available at http://localhost:8000${NC}"

down: ## Stop all development services
	@echo "${YELLOW}Stopping development services...${NC}"
	docker-compose -f $(COMPOSE_FILE) down

logs: ## View logs for all services
	docker-compose -f $(COMPOSE_FILE) logs -f

shell: ## Open shell in API container
	docker-compose -f $(COMPOSE_FILE) exec api /bin/bash

# Production commands
prod-build: ## Build all services for production
	@echo "${YELLOW}Building production services...${NC}"
	docker-compose -f $(PROD_COMPOSE_FILE) build

prod-up: ## Start all services in production mode
	@echo "${GREEN}Starting production services...${NC}"
	docker-compose -f $(PROD_COMPOSE_FILE) --env-file $(ENV_FILE) up -d
	@echo "${GREEN}Production services started!${NC}"

prod-down: ## Stop all production services
	@echo "${YELLOW}Stopping production services...${NC}"
	docker-compose -f $(PROD_COMPOSE_FILE) down

prod-logs: ## View production logs
	docker-compose -f $(PROD_COMPOSE_FILE) logs -f

# Utility commands
clean: ## Clean up containers, volumes, and images
	@echo "${RED}Cleaning up Docker resources...${NC}"
	docker-compose -f $(COMPOSE_FILE) down -v --rmi local
	docker-compose -f $(PROD_COMPOSE_FILE) down -v --rmi local

test: ## Run tests in Docker
	@echo "${YELLOW}Running tests...${NC}"
	docker-compose -f $(COMPOSE_FILE) run --rm api pytest

health: ## Check health of all services
	@echo "${YELLOW}Checking service health...${NC}"
	@docker-compose -f $(COMPOSE_FILE) ps
	@echo ""
	@echo "${YELLOW}API Health:${NC}"
	@curl -s http://localhost:8000/health | jq . || echo "${RED}API not responding${NC}"

backup: ## Backup data volumes
	@echo "${YELLOW}Backing up data...${NC}"
	@mkdir -p backups
	@docker run --rm -v virtuoso_redis-data:/data -v $(PWD)/backups:/backup alpine tar czf /backup/redis-backup-$$(date +%Y%m%d-%H%M%S).tar.gz -C /data .
	@echo "${GREEN}Backup completed!${NC}"

restore: ## Restore data from backup (requires BACKUP_FILE)
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "${RED}Error: BACKUP_FILE not specified${NC}"; \
		echo "Usage: make restore BACKUP_FILE=backups/redis-backup-YYYYMMDD-HHMMSS.tar.gz"; \
		exit 1; \
	fi
	@echo "${YELLOW}Restoring from $(BACKUP_FILE)...${NC}"
	@docker run --rm -v virtuoso_redis-data:/data -v $(PWD)/$(BACKUP_FILE):/backup.tar.gz alpine tar xzf /backup.tar.gz -C /data
	@echo "${GREEN}Restore completed!${NC}"

# Development helpers
cli-test: ## Test CLI commands
	@echo "${YELLOW}Testing CLI...${NC}"
	docker-compose -f $(COMPOSE_FILE) run --rm cli list-commands

api-shell: ## Open Python shell in API container
	docker-compose -f $(COMPOSE_FILE) exec api python

redis-cli: ## Open Redis CLI
	docker-compose -f $(COMPOSE_FILE) exec redis redis-cli

# Monitoring
stats: ## Show container statistics
	docker stats --no-stream $$(docker-compose -f $(COMPOSE_FILE) ps -q)

# Installation
install: ## Initial setup
	@echo "${YELLOW}Setting up Virtuoso Docker environment...${NC}"
	@cp -n .env.example .env || true
	@echo "${GREEN}Please edit .env file with your configuration${NC}"
	@echo "${GREEN}Then run 'make build' and 'make up'${NC}"
