#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting Blogging Platform...${NC}"
echo ""

# Kill any existing processes on ports 8080 and 5173
echo -e "${YELLOW}Cleaning up ports...${NC}"
for PORT in 8080 5173; do
    if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        PID=$(lsof -Pi :$PORT -sTCP:LISTEN -t)
        echo -e "${YELLOW}Killing process on port $PORT (PID: $PID)...${NC}"
        kill -9 $PID 2>/dev/null || true
    fi
done
echo -e "${GREEN}✓ Ports cleaned${NC}"
echo ""

# Check if PostgreSQL is running
echo -e "${YELLOW}Checking PostgreSQL...${NC}"
if ! docker ps | grep -q takehome-kshitij-dhara-postgres; then
    echo -e "${YELLOW}PostgreSQL not running. Starting Docker Compose...${NC}"
    docker compose up -d
    sleep 3
fi
echo -e "${GREEN}✓ PostgreSQL is running${NC}"
echo ""

# Start Frontend
echo -e "${BLUE}Starting Frontend (Vite Dev Server) on :5173...${NC}"
cd frontend
npm run dev > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..
echo -e "${GREEN}✓ Frontend started (PID: $FRONTEND_PID)${NC}"
sleep 2

echo -e "${GREEN}✓ All services running!${NC}"
echo ""

echo "==============================================="
echo "  🚀 Application URLs:"
echo "  Frontend:  http://localhost:5173"
echo "  Backend:   http://localhost:8080"
echo "  Health:    http://localhost:8080/api/v1/health"
echo ""
echo "  Frontend logs: /tmp/frontend.log"
echo "  Ctrl+C to stop all services"
echo "==============================================="
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down...${NC}"
    kill $FRONTEND_PID 2>/dev/null || true
    echo -e "${GREEN}✓ Services stopped${NC}"
    exit 0
}

# Set trap to cleanup on Ctrl+C
trap cleanup SIGINT SIGTERM

# Start Backend in foreground (will show all logs)
echo -e "${BLUE}Starting Backend (Go) on :8080...${NC}"
cd backend
go run cmd/server/main.go
