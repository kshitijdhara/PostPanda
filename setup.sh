#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Blogging Platform Setup${NC}"
echo "=================================="
echo ""

# Track if any installation happened
NEEDS_RESTART=false

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        return 1
    fi
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check Go
echo -e "${BLUE}Checking Go installation...${NC}"
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_status 0 "Go is installed ($GO_VERSION)"
else
    print_warning "Go is not installed"
    echo "  Install from: https://golang.org/doc/install"
    echo "  Or use: brew install go (macOS)"
    echo ""
    exit 1
fi
echo ""

# Check Node.js and npm
echo -e "${BLUE}Checking Node.js and npm...${NC}"
if command_exists node && command_exists npm; then
    NODE_VERSION=$(node --version)
    NPM_VERSION=$(npm --version)
    print_status 0 "Node.js is installed ($NODE_VERSION)"
    print_status 0 "npm is installed ($NPM_VERSION)"
else
    print_warning "Node.js or npm is not installed"
    echo "  Install from: https://nodejs.org/"
    echo "  Or use: brew install node (macOS)"
    echo ""
    exit 1
fi
echo ""

# Check Docker
echo -e "${BLUE}Checking Docker...${NC}"
if command_exists docker; then
    DOCKER_VERSION=$(docker --version)
    print_status 0 "$DOCKER_VERSION"
else
    print_warning "Docker is not installed"
    echo "  Install from: https://www.docker.com/products/docker-desktop"
    echo "  Or use: brew install docker (macOS, requires Docker Desktop)"
    echo ""
    exit 1
fi
echo ""

# Check Docker Compose
echo -e "${BLUE}Checking Docker Compose...${NC}"
if command_exists docker-compose; then
    DC_VERSION=$(docker-compose --version)
    print_status 0 "$DC_VERSION"
elif docker compose version >/dev/null 2>&1; then
    print_status 0 "Docker Compose (integrated)"
else
    print_warning "Docker Compose is not installed"
    echo "  Install from: https://docs.docker.com/compose/install/"
    echo ""
    exit 1
fi
echo ""

# Check PostgreSQL client (optional but helpful)
echo -e "${BLUE}Checking PostgreSQL client...${NC}"
if command_exists psql; then
    PSQL_VERSION=$(psql --version)
    print_status 0 "$PSQL_VERSION"
else
    print_warning "PostgreSQL client is not installed (optional)"
    echo "  Install from: https://www.postgresql.org/download/"
    echo "  Or use: brew install postgresql (macOS)"
fi
echo ""

# Setup backend
echo -e "${BLUE}Setting up Backend...${NC}"
if [ -d "backend" ]; then
    cd backend

    # Check if go.mod exists
    if [ ! -f "go.mod" ]; then
        print_warning "go.mod not found, initializing Go module"
        go mod init github.com/kshitijdhara/blog 2>/dev/null || true
    fi

    # Download dependencies
    echo "  Downloading Go dependencies..."
    go mod download
    go mod tidy
    print_status 0 "Backend dependencies downloaded"

    cd ..
else
    echo -e "${RED}✗ backend directory not found${NC}"
    exit 1
fi
echo ""

# Setup frontend
echo -e "${BLUE}Setting up Frontend...${NC}"
if [ -d "frontend" ]; then
    cd frontend

    # Install npm dependencies
    echo "  Installing npm dependencies (this may take a moment)..."
    npm ci
    print_status 0 "Frontend dependencies installed"

    cd ..
else
    echo -e "${RED}✗ frontend directory not found${NC}"
    exit 1
fi
echo ""

# Create .env file if it doesn't exist
echo -e "${BLUE}Configuring environment...${NC}"
if [ ! -f ".env" ]; then
    print_status 0 "Creating .env file"
    cat .env.example > .env
    echo "  Updated .env with defaults from .env.example"
else
    print_status 0 ".env file already exists"
fi
echo ""

# Check if Docker daemon is running
echo -e "${BLUE}Checking Docker daemon...${NC}"
if docker ps >/dev/null 2>&1; then
    print_status 0 "Docker daemon is running"
else
    print_warning "Docker daemon is not running"
    echo "  Start Docker Desktop or daemon before running services"
fi
echo ""

# Make scripts executable
echo -e "${BLUE}Making scripts executable...${NC}"
chmod +x build.sh start.sh setup.sh 2>/dev/null
print_status 0 "Scripts are executable"
echo ""

# Final summary
echo "=================================="
echo -e "${GREEN}✓ Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review and update .env if needed"
echo "  2. Start PostgreSQL:  docker compose up -d"
echo "  3. Build the apps:    ./build.sh"
echo "  4. Run the apps:      ./start.sh"
echo ""
echo "Documentation:"
echo "  - README.md            (project overview)"
echo "  - .env.example         (environment variables)"
echo "  - docker-compose.yml   (infrastructure)"
echo ""
