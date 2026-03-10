#!/bin/bash

set -e

echo "🔨 Building Blogging Platform..."
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build Backend
echo -e "${BLUE}Building Backend (Go)...${NC}"
cd backend
go mod download
go build -o server cmd/server/main.go
echo -e "${GREEN}✓ Backend built successfully${NC}"
cd ..

echo ""

# Build Frontend
echo -e "${BLUE}Building Frontend (React/Vite)...${NC}"
cd frontend
npm ci
npm run build
echo -e "${GREEN}✓ Frontend built successfully${NC}"
cd ..

echo ""
echo -e "${GREEN}✓ Build complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Start PostgreSQL: docker compose up -d"
echo "  2. Run: ./start.sh"
