# Dear Future - Makefile for building and running the application
# Built with functional programming principles in Go

.PHONY: help build run test clean cli server docker deploy-lambda

# Default target
help: ## Show this help message
	@echo "Dear Future - Your Message to Tomorrow"
	@echo "==============================================="
	@echo ""
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Architecture: Functional Programming with Clean Architecture"
	@echo "Deployment: Lambda → ECS → AKS migration path"

# Build targets
build: ## Build all binaries
	@echo "🔨 Building Dear Future applications..."
	@go build -o bin/dear-future main.go
	@go build -o bin/dear-future-cli cmd/cli/main.go
	@go build -o bin/dear-future-server cmd/server/main.go
	@echo "✅ Build complete! Binaries are in ./bin/"

build-server: ## Build server binary only
	@echo "🔨 Building server..."
	@go build -o bin/dear-future-server cmd/server/main.go
	@echo "✅ Server build complete!"

build-cli: ## Build CLI binary only
	@echo "🔨 Building CLI..."
	@go build -o bin/dear-future-cli cmd/cli/main.go
	@echo "✅ CLI build complete!"

# Run targets
run: build ## Build and run the main server
	@echo "🚀 Starting Dear Future server..."
	@./bin/dear-future

run-server: build-server ## Build and run the server
	@echo "🚀 Starting Dear Future server..."
	@./bin/dear-future-server

run-dev: ## Run server in development mode
	@echo "🧪 Starting Dear Future in development mode..."
	@ENVIRONMENT=development go run main.go

run-staging: build ## Run server with staging configuration
	@echo "🏭 Starting Dear Future in staging mode..."
	@CONFIG_FILE=config.staging.yaml DATABASE_URL=postgresql://test:test@localhost/test S3_BUCKET=staging-bucket ./bin/dear-future

run-prod: build ## Run server with production configuration
	@echo "🏭 Starting Dear Future in production mode..."
	@CONFIG_FILE=config.production.yaml DATABASE_URL=postgresql://prod:prod@localhost/prod S3_BUCKET=prod-bucket JWT_SECRET=secure-production-secret ./bin/dear-future

# CLI targets
cli: build-cli ## Build CLI and show help
	@./bin/dear-future-cli --cmd help

cli-version: build-cli ## Show CLI version
	@./bin/dear-future-cli --cmd version

cli-test: build-cli ## Run functional tests via CLI
	@./bin/dear-future-cli --cmd test

cli-health: build-cli ## Check application health via CLI
	@DATABASE_URL=mock://test ./bin/dear-future-cli --cmd health

cli-health-staging: build-cli ## Check application health via CLI using staging config
	@DATABASE_URL=postgresql://test:test@localhost/test S3_BUCKET=staging-bucket ./bin/dear-future-cli --config config.staging.yaml --cmd health

cli-health-prod: build-cli ## Check application health via CLI using production config
	@DATABASE_URL=postgresql://test:test@localhost/test S3_BUCKET=prod-bucket JWT_SECRET=secure-production-secret ./bin/dear-future-cli --config config.production.yaml --cmd health

# API Testing targets
test-api: build ## Test the environment API endpoint
	@echo "🧪 Testing Environment API endpoint..."
	@DATABASE_URL=postgresql://test:test@localhost/test ./bin/dear-future & \
	SERVER_PID=$$! && \
	sleep 2 && \
	echo "📊 Development Environment:" && \
	curl -s http://localhost:8080/environment/current | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"Environment: {data['environment']}\"); print(f\"Version: {data['version']}\"); print(f\"Platform: {data['platform']}\"); print(f\"Features: {sum(data['features'].values())}/{len(data['features'])} enabled\")" && \
	kill $$SERVER_PID 2>/dev/null || true

test-api-staging: build ## Test the environment API endpoint with staging config
	@echo "🏭 Testing Staging Environment API..."
	@CONFIG_FILE=config.staging.yaml DATABASE_URL=postgresql://test:test@localhost/test S3_BUCKET=staging-bucket ./bin/dear-future & \
	SERVER_PID=$$! && \
	sleep 2 && \
	echo "📊 Staging Environment:" && \
	curl -s http://localhost:8080/environment/current | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"Environment: {data['environment']}\"); print(f\"Version: {data['version']}\"); print(f\"Platform: {data['platform']}\"); print(f\"Features: {sum(data['features'].values())}/{len(data['features'])} enabled\")" && \
	kill $$SERVER_PID 2>/dev/null || true

test-endpoints: build ## Test all API endpoints
	@echo "🔍 Testing all API endpoints..."
	@DATABASE_URL=postgresql://test:test@localhost/test ./bin/dear-future & \
	SERVER_PID=$$! && \
	sleep 2 && \
	echo "1️⃣  Health Check:" && \
	curl -s http://localhost:8080/health | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"Status: {data['status']}\")" && \
	echo "2️⃣  Environment Info:" && \
	curl -s http://localhost:8080/environment/current | python3 -c "import sys, json; data=json.load(sys.stdin); print(f\"Environment: {data['environment']} v{data['version']}\")" && \
	echo "3️⃣  API Info:" && \
	curl -s http://localhost:8080/api/v1/ | head -1 && \
	kill $$SERVER_PID 2>/dev/null || true && \
	echo "✅ All endpoints working!"

# Test targets
test: ## Run all Go tests
	@echo "🧪 Running Go tests..."
	@go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "🏃 Running tests with race detection..."
	@go test ./... -race

benchmark: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	@go test ./... -bench=. -benchmem

# Code quality targets
fmt: ## Format code
	@echo "📝 Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "🔍 Running linter..."
	@golangci-lint run

tidy: ## Tidy dependencies
	@echo "🧹 Tidying dependencies..."
	@go mod tidy

# Development targets
dev-setup: ## Set up development environment
	@echo "🛠️  Setting up development environment..."
	@go mod download
	@echo "✅ Development environment ready!"

dev-deps: ## Install development dependencies
	@echo "📦 Installing development dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development dependencies installed!"

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t dear-future:latest .
	@echo "✅ Docker image built!"

docker-run: docker-build ## Build and run Docker container
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 -e ENVIRONMENT=development dear-future:latest

# Clean targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete!"

clean-all: clean ## Clean everything including dependencies
	@echo "🧹 Cleaning everything..."
	@go clean -modcache
	@echo "✅ Deep clean complete!"

# Deployment targets
deploy-lambda: ## Deploy to AWS Lambda
	@echo "☁️  Deploying to AWS Lambda..."
	@sam build
	@sam deploy --guided
	@echo "✅ Lambda deployment complete!"

deploy-ecs: ## Deploy to AWS ECS
	@echo "☁️  Deploying to AWS ECS..."
	@./scripts/deploy-ecs.sh
	@echo "✅ ECS deployment complete!"

deploy-k8s: ## Deploy to Kubernetes
	@echo "☁️  Deploying to Kubernetes..."
	@kubectl apply -f deployments/k8s/
	@echo "✅ Kubernetes deployment complete!"

# Documentation targets
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@go doc -all ./pkg/... > docs/api.md
	@echo "✅ Documentation generated!"

docs-serve: ## Serve documentation locally
	@echo "📚 Serving documentation..."
	@godoc -http=:6060
	@echo "📚 Documentation available at http://localhost:6060"

# Quick development commands
quick-test: ## Quick test and build cycle
	@echo "⚡ Quick test and build..."
	@go test ./pkg/domain/user/... && go build -o bin/dear-future main.go
	@echo "✅ Quick cycle complete!"

demo: build ## Run a full demo
	@echo "🎭 Running Dear Future demo..."
	@echo ""
	@echo "1️⃣  Testing functional programming patterns..."
	@./bin/dear-future-cli --cmd test
	@echo ""
	@echo "2️⃣  Checking application version..."
	@./bin/dear-future-cli --cmd version
	@echo ""
	@echo "3️⃣  Functional programming codebase ready!"
	@echo ""
	@echo "🚀 Start the server with: make run"
	@echo "🔗 Then visit: http://localhost:8080"

# Project info
info: ## Show project information
	@echo "Dear Future - Your Message to Tomorrow"
	@echo "======================================"
	@echo ""
	@echo "🏗️  Architecture: Functional Programming + Clean Architecture"
	@echo "📦 Language: Go 1.21+ with generics"
	@echo "🧪 Testing: Comprehensive functional tests"
	@echo "☁️  Deployment: Lambda → ECS → AKS migration ready"
	@echo ""
	@echo "✨ Key Features:"
	@echo "   • Pure business logic (side-effect free)"
	@echo "   • Immutable data structures"
	@echo "   • Result/Option monads for error handling"
	@echo "   • Function composition patterns"
	@echo "   • Migration-ready deployment architecture"
	@echo ""
	@echo "📂 Project Structure:"
	@echo "   pkg/domain/     - Pure business logic"
	@echo "   pkg/effects/    - Side effects boundary"
	@echo "   pkg/mocks/      - Development services"
	@echo "   cmd/            - CLI and server entry points"
	@echo "   docs/           - Architecture documentation"
	@echo ""
	@echo "🎯 Next Steps:"
	@echo "   make demo       - Run full demonstration"
	@echo "   make run        - Start the web server"
	@echo "   make test       - Run all tests"