#!/bin/bash

# Development startup script for Dear Future
# Starts PostgreSQL, Backend, and Frontend with hot reload

set -e

echo "🚀 Starting Dear Future Development Environment"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    exit 1
fi

# Start PostgreSQL and Adminer
echo -e "${BLUE}📦 Starting PostgreSQL and Adminer...${NC}"
docker-compose up -d postgres adminer

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}⏳ Waiting for PostgreSQL to be ready...${NC}"
until docker exec dear-future-postgres pg_isready -U postgres > /dev/null 2>&1; do
    sleep 1
done
echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found, copying from .env.local${NC}"
    cp .env.local .env
fi

# Check if node_modules exists in ui/
if [ ! -d "ui/node_modules" ]; then
    echo -e "${BLUE}📦 Installing UI dependencies...${NC}"
    cd ui && npm install && cd ..
    echo ""
fi

echo -e "${GREEN}✅ Development environment ready!${NC}"
echo ""
echo "Services:"
echo -e "  ${BLUE}Backend:${NC}  http://localhost:8080"
echo -e "  ${BLUE}Frontend:${NC} http://localhost:3000"
echo -e "  ${BLUE}Adminer:${NC}  http://localhost:8081"
echo ""
echo -e "${YELLOW}Starting Backend and Frontend servers...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}⏹️  Stopping services...${NC}"
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    echo -e "${GREEN}✅ Services stopped${NC}"
    exit 0
}

trap cleanup INT TERM

# Start backend in background
echo -e "${BLUE}[Backend]${NC} Starting Go server..."
go run main.go &
BACKEND_PID=$!

# Wait a bit for backend to start
sleep 2

# Start frontend in background
echo -e "${BLUE}[Frontend]${NC} Starting Next.js dev server..."
cd ui && npm run dev &
FRONTEND_PID=$!
cd ..

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
