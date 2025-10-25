# Dear Future - Your Message to Tomorrow

A web application for sending messages to your future self, built with functional programming principles and designed for ultra-low cost scaling.

## 🎯 Project Overview

Dear Future allows users to schedule messages to be delivered to themselves at future dates. Whether it's a reminder for next year, encouragement for tough times, or just a note to your future self, this app makes it easy to communicate across time.

## ✨ Key Features

- **Schedule Messages**: Send messages to yourself at any future date
- **Timezone Support**: Accurate delivery regardless of timezone changes
- **File Attachments**: Include photos, documents, or other files
- **User Profiles**: Manage delivery preferences and notification settings
- **Email Delivery**: Reliable message delivery via email
- **Smart Scheduling**: Optimal delivery times based on user preferences

## 🏗️ Architecture Highlights

### Functional Programming Design
- **Pure Business Logic**: All domain logic is side-effect free
- **Immutable Data Structures**: Data cannot be accidentally modified
- **Result/Option Monads**: Elegant error handling without exceptions
- **Function Composition**: Complex operations built from simple functions
- **Side Effect Isolation**: I/O operations clearly separated from business logic

### Migration-Ready Deployment
- **Start Ultra-Cheap**: AWS Lambda ($0-5/month for 1-10K users)
- **Scale Efficiently**: AWS ECS ($30-50/month for 10K-100K users)
- **Enterprise Ready**: Azure AKS ($100-300/month for 100K+ users)
- **Same Codebase**: No rewrites needed when scaling between platforms

## 🛠️ Tech Stack

### Backend (Implemented)
- **Language**: Go 1.21+ with functional programming patterns
- **Database**: PostgreSQL (local via Docker) + Supabase (production)
- **Storage**: Cloudflare R2 (S3-compatible, zero egress fees)
- **Email**: Gmail SMTP (free for 500 emails/day) or SendGrid
- **Authentication**: Supabase Auth (planned)
- **Deployment**: AWS Lambda → ECS → AKS

### Frontend (Planned)
- **Framework**: React 18 with TypeScript
- **Build Tool**: Vite for fast development
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Server State**: React Query
- **Deployment**: Vercel

## 📁 Project Structure

```
dear-future/
├── docs/                           # Comprehensive documentation
│   ├── architecture.md            # System architecture overview
│   ├── functional-programming-guide.md # FP implementation guide
│   ├── deployment-guide.md         # Multi-platform deployment
│   └── project-structure.md        # Monorepo organization
├── pkg/                           # Go application code
│   ├── domain/                    # Pure business logic
│   │   ├── common/               # Functional programming utilities
│   │   │   ├── result.go         # Result monad implementation
│   │   │   ├── option.go         # Option monad implementation
│   │   │   └── functional.go     # Higher-order functions
│   │   ├── user/                 # User domain
│   │   │   ├── types.go          # Immutable user types
│   │   │   ├── validation.go     # Pure validation functions
│   │   │   └── user_test.go      # Comprehensive tests
│   │   └── message/              # Message domain
│   │       ├── types.go          # Immutable message types
│   │       ├── validation.go     # Message validation
│   │       └── business.go       # Business logic functions
│   ├── effects/                   # Side effects boundary
│   │   └── interfaces.go         # I/O operation interfaces
│   ├── config/                   # Configuration management
│   │   └── config.go             # Environment-based config
│   └── composition/              # Dependency injection
│       └── app.go                # Application composition
├── ui/                           # React frontend (planned)
└── deployments/                  # Deployment configurations
    ├── lambda/                   # AWS Lambda deployment
    ├── ecs/                      # AWS ECS deployment
    └── k8s/                      # Kubernetes deployment
```

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose (for local database)
- Node.js 18+ (for frontend, when implemented)
- AWS CLI configured (for deployment)
- Supabase account (for production database)

### Local Development Setup

For local development, you can use PostgreSQL running in Docker instead of Supabase:

1. **Start the local PostgreSQL database**
```bash
docker-compose up -d
```

This will start:
- PostgreSQL 16 on port 5432
- Adminer (database UI) on port 8081 (optional, access at http://localhost:8081)

2. **Copy the local environment configuration**
```bash
cp .env.local .env
```

The local configuration uses:
- Database: `postgresql://postgres:postgres@localhost:5432/dear_future_dev?sslmode=disable`
- Mock services for AWS S3 and SES (no AWS account needed)
- Debug mode enabled

3. **Apply database migrations**

The migrations will be automatically applied when the database starts for the first time. You can also manually apply them:

```bash
docker exec -i dear-future-postgres psql -U postgres -d dear_future_dev < migrations/001_init_schema.sql
```

4. **Access the database**

Using Adminer (web UI):
- URL: http://localhost:8081
- System: PostgreSQL
- Server: postgres
- Username: postgres
- Password: postgres
- Database: dear_future_dev

Or using psql directly:
```bash
docker exec -it dear-future-postgres psql -U postgres -d dear_future_dev
```

5. **Stop the local database**
```bash
docker-compose down
```

To remove all data and start fresh:
```bash
docker-compose down -v
```

### Running with Hot Reload (Full Stack)

Now you can run both backend and frontend with hot reload enabled:

**Option 1: Run everything together (Recommended)**
```bash
make dev
```

This single command will:
- Start PostgreSQL and Adminer
- Start the Go backend with auto-reload
- Start the Next.js frontend with hot reload (Turbopack)
- Configure API proxying automatically

Access:
- Frontend: http://localhost:3000 (with hot reload)
- Backend API: http://localhost:8080
- Adminer: http://localhost:8081

Press `Ctrl+C` to stop all services.

**Option 2: Run services separately**

In separate terminal windows:

```bash
# Terminal 1: Start database
make db-start

# Terminal 2: Start backend
make dev-backend

# Terminal 3: Start UI with hot reload
make dev-ui
```

**Option 3: Manual control**

```bash
# Database
docker-compose up -d postgres adminer

# Backend
go run main.go

# UI (in the ui/ directory)
cd ui && npm run dev
```

### Hot Reload Features

- **Frontend (Next.js)**: Uses Turbopack for instant hot module replacement (HMR)
- **Backend (Go)**: Restart `go run main.go` when you make changes, or use a tool like `air` for auto-reload
- **Database**: Changes persist in Docker volume, survives restarts

### API Proxy Configuration

The UI is configured to proxy API requests to the backend automatically:
- Requests to `http://localhost:3000/api/*` are forwarded to `http://localhost:8080/api/*`
- This avoids CORS issues during development
- Configuration is in [ui/next.config.ts](ui/next.config.ts)

### Production Setup (Supabase)

For production deployment, use Supabase instead of local PostgreSQL:

### Environment Setup

1. **Clone the repository**
```bash
git clone https://github.com/your-username/dear-future.git
cd dear-future
```

2. **Install Go dependencies**
```bash
go mod download
```

3. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

Required environment variables:
```bash
ENVIRONMENT=development
DATABASE_URL=postgresql://...  # Supabase connection string
JWT_SECRET=your-secret-key
AWS_REGION=us-east-1
S3_BUCKET=your-bucket
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_KEY=your-service-key
```

### Running the Application

**Development mode with hot reload (full stack):**
```bash
make dev
```
This starts PostgreSQL, backend, and frontend with hot reload enabled.

**Run backend only:**
```bash
make dev-backend
```

**Run UI only (with hot reload):**
```bash
make dev-ui
```

**Run tests:**
```bash
go test ./... -v
```

**Build for production:**
```bash
go build -o dear-future main.go
```

## 🧪 Testing

The codebase includes comprehensive tests demonstrating functional programming concepts:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run user domain tests (with detailed output)
go test ./pkg/domain/user/... -v

# Run benchmarks
go test ./pkg/domain/user/... -bench=.
```

### Test Coverage
- ✅ Pure function testing
- ✅ Immutability verification
- ✅ Error handling with Result types
- ✅ Data validation
- ✅ Business logic correctness
- ✅ Function composition

## 📊 Functional Programming Examples

### Result Monad for Error Handling
```go
// Instead of returning (value, error), we use Result[T]
func CreateUser(req CreateUserRequest) common.Result[User] {
    return common.Bind(
        validateUserRequest(req),
        func(validReq CreateUserRequest) common.Result[User] {
            return createUserFromRequest(validReq)
        },
    )
}
```

### Immutable Data Structures
```go
// Users are immutable - updates return new instances
user := NewUser(req).Value()
updatedUser := user.WithName("New Name").Value()
// Original user is unchanged
```

### Pure Business Logic
```go
// No side effects, deterministic output
func CalculateOptimalDeliveryTime(requested time.Time, timezone string) Result[time.Time] {
    // Pure calculation logic
}
```

## 🚢 Deployment Guide

### Phase 1: Lambda (Start Here - $0-5/month)
```bash
sam build
sam deploy --guided
```

### Phase 2: ECS (Scale Up - $30-50/month)
```bash
docker build -t dear-future .
./scripts/deploy-ecs.sh
```

### Phase 3: AKS (Enterprise - $100-300/month)
```bash
kubectl apply -f deployments/k8s/
```

## 📚 Documentation

- **[Architecture Guide](docs/architecture.md)** - Complete system design
- **[Infrastructure Setup](docs/infrastructure-setup.md)** - PostgreSQL, R2, and SMTP configuration
- **[API Implementation](docs/api-implementation.md)** - REST API endpoints and authentication
- **[Functional Programming Guide](docs/functional-programming-guide.md)** - FP patterns and practices
- **[Deployment Guide](docs/deployment-guide.md)** - Multi-platform deployment strategies
- **[Project Structure](docs/project-structure.md)** - Monorepo organization

## 🎯 Development Roadmap

### Phase 1: Backend Foundation ✅
- [x] Functional programming architecture
- [x] Domain models (User, Message)
- [x] Business logic functions
- [x] Configuration management
- [x] Dependency injection
- [x] Comprehensive tests

### Phase 2: Infrastructure & API ✅ COMPLETE
- [x] PostgreSQL database adapter
- [x] Cloudflare R2 storage implementation (S3-compatible)
- [x] Gmail SMTP email service
- [x] JWT authentication system
- [x] Authentication & security middleware
- [x] User API handlers (register, login, profile)
- [x] Message API handlers (CRUD operations)
- [x] Complete REST API with 12 endpoints

### Phase 3: Frontend Development
- [ ] React TypeScript application
- [ ] User authentication flow
- [ ] Message creation/management
- [ ] File upload interface
- [ ] Responsive design

### Phase 4: Production Features
- [ ] Message scheduling system
- [ ] Email templates
- [ ] User dashboard
- [ ] Analytics and monitoring
- [ ] Advanced scheduling options

## 🔧 Configuration

The application supports extensive configuration through environment variables:

### Core Settings
- `ENVIRONMENT`: development, staging, production
- `PORT`: Server port (default: 8080)
- `DEBUG`: Enable debug logging

### Database
- `DATABASE_URL`: PostgreSQL connection string
- `SUPABASE_URL`: Supabase project URL
- `SUPABASE_SERVICE_KEY`: Supabase service role key

### AWS Services
- `AWS_REGION`: AWS region for services
- `S3_BUCKET`: S3 bucket for file storage
- `SES_FROM_EMAIL`: Email address for sending messages

### Feature Flags
- `FEATURE_FILE_ATTACHMENTS`: Enable file attachments
- `FEATURE_BATCH_PROCESSING`: Enable batch operations
- `FEATURE_ANALYTICS`: Enable usage analytics

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow functional programming principles
4. Write comprehensive tests
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Guidelines
- All business logic must be pure functions
- Use immutable data structures
- Handle errors with Result/Option types
- Write tests for all public functions
- Document complex business logic

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by functional programming languages like Haskell and F#
- Built with Go's excellent tooling and ecosystem
- Designed for cost-effective scaling on cloud platforms

---

**Built with ❤️ and functional programming principles**

*Send a message to your future self today!*
