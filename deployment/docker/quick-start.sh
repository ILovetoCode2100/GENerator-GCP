#!/bin/bash
# Quick start script for Virtuoso Docker setup

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Virtuoso API CLI - Docker Quick Start${NC}"
echo "======================================"
echo ""

# Check Docker installation
echo -n "Checking Docker installation... "
if command -v docker &> /dev/null && docker --version &> /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo -e "${RED}Error: Docker is not installed or not running${NC}"
    echo "Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Docker Compose
echo -n "Checking Docker Compose... "
if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

# Create .env file if not exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from template...${NC}"
    cp .env.example .env
    echo -e "${YELLOW}Please edit .env file with your Virtuoso API credentials${NC}"
    echo "Required variables:"
    echo "  - VIRTUOSO_API_KEY"
    echo "  - VIRTUOSO_ORG_ID"
    echo ""
    read -p "Press Enter after updating .env file..."
fi

# Build or pull images
echo -e "${YELLOW}Building Docker images...${NC}"
docker-compose -f docker-compose.prod.yml build

# Start services
echo -e "${GREEN}Starting services...${NC}"
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 10

# Check health
echo -e "${YELLOW}Checking service health...${NC}"
if curl -f http://localhost:8000/health &> /dev/null; then
    echo -e "${GREEN}✓ API is healthy${NC}"
else
    echo -e "${RED}✗ API health check failed${NC}"
    echo "Check logs with: docker-compose -f docker-compose.prod.yml logs api"
fi

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Services running:"
echo "  - API: http://localhost:8000"
echo "  - API Docs: http://localhost:8000/docs"
echo ""
echo "Useful commands:"
echo "  - View logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "  - Stop services: docker-compose -f docker-compose.prod.yml down"
echo "  - Run CLI: docker-compose -f docker-compose.prod.yml run cli <command>"
echo ""
echo "For more information, see README.md"
