# Quick Start Guide

Get up and running with Dear Future in 2 minutes!

## Prerequisites

- Docker Desktop installed and running
- Go 1.21+ installed
- Node.js 18+ installed

## Quick Start (3 steps)

### 1. Set up your environment

```bash
# Clone and enter the project
git clone https://github.com/your-username/dear-future.git
cd dear-future

# Install all dependencies
make dev-setup
```

### 2. Start everything

```bash
make dev
```

This will start:
- PostgreSQL database (port 5432)
- Backend API (port 8080)
- Frontend with hot reload (port 3000)
- Database UI (port 8081)

### 3. Open your browser

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1/
- **Health Check**: http://localhost:8080/health
- **Database UI**: http://localhost:8081

## Development Workflow

### Running Services Separately

If you prefer more control, run services in separate terminals:

```bash
# Terminal 1: Database
make db-start

# Terminal 2: Backend
make dev-backend

# Terminal 3: Frontend (with hot reload)
make dev-ui
```

### Making Changes

**Frontend Changes** (React/Next.js in `ui/`):
- Edit files in `ui/src/`
- Changes are instantly reflected with Turbopack hot reload
- No server restart needed

**Backend Changes** (Go in `pkg/`):
- Edit files in `pkg/`
- Restart the backend: `Ctrl+C` and run `make dev-backend` again
- Or use [air](https://github.com/cosmtrek/air) for auto-reload

**Database Changes**:
- Create new migration files in `migrations/`
- Apply manually: `make db-reset` (warning: clears data)

### Useful Commands

```bash
# See all available commands
make help

# Run tests
make test

# Format code
make fmt

# Stop all services
make db-stop

# Reset database (clears all data)
make db-reset
```

## Database Access

### Using Adminer (Web UI)

Go to http://localhost:8081 and login with:
- System: PostgreSQL
- Server: postgres
- Username: postgres
- Password: postgres
- Database: dear_future_dev

### Using psql (Command Line)

```bash
docker exec -it dear-future-postgres psql -U postgres -d dear_future_dev
```

## Troubleshooting

### "Docker is not running"
- Start Docker Desktop
- Run `make dev` again

### "Port already in use"
- Stop other services using ports 3000, 5432, 8080, or 8081
- Or change ports in configuration files

### "Database connection failed"
- Check if PostgreSQL is running: `docker ps | grep postgres`
- Restart database: `make db-stop && make db-start`

### "Module not found" in UI
- Install UI dependencies: `cd ui && npm install`

### Backend won't start
- Check .env file exists: `ls -la .env`
- Copy from template: `cp .env.local .env`

## Environment Variables

### Local Development (.env)
```bash
ENVIRONMENT=development
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/dear_future_dev?sslmode=disable
PORT=8080
DEBUG=true
```

### Production (Supabase)
```bash
ENVIRONMENT=production
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_KEY=your-service-key
DATABASE_URL=postgresql://user:pass@host:5432/db
JWT_SECRET=your-secure-secret
```

## Project Structure

```
dear-future/
├── pkg/              # Go backend code
│   ├── domain/       # Business logic (pure functions)
│   ├── effects/      # Side effects (database, API)
│   ├── config/       # Configuration management
│   └── server/       # HTTP server
├── ui/               # Next.js frontend
│   └── src/app/      # React components
├── migrations/       # Database migrations
├── scripts/          # Helper scripts
└── Makefile          # Development commands
```

## Next Steps

1. **Explore the codebase**: Check out [docs/architecture.md](docs/architecture.md)
2. **Read the full README**: See [README.md](README.md)
3. **Run tests**: `make test`
4. **Start building**: Create your first feature!

## Need Help?

- Full documentation: [README.md](README.md)
- Architecture guide: [docs/architecture.md](docs/architecture.md)
- Configuration guide: [docs/configuration-guide.md](docs/configuration-guide.md)

---

**Happy coding!** 🚀
