# Project Structure - Monorepo Architecture

## Overview

The Dear Future project uses a monorepo structure with the Go backend and React frontend in the same repository. This approach provides better code organization, shared tooling, and simplified development workflow.

## Complete Project Structure

```
dear-future/
├── README.md
├── go.mod                          # Go module file
├── go.sum                          # Go dependencies
├── Makefile                        # Build automation
├── .gitignore
├── .github/
│   └── workflows/
│       ├── backend.yml             # Backend CI/CD
│       ├── frontend.yml            # Frontend CI/CD
│       └── deploy.yml              # Deployment pipeline
├── docs/                           # Architecture documentation
│   ├── architecture.md
│   ├── functional-programming-guide.md
│   ├── deployment-guide.md
│   └── project-structure.md
├── scripts/                        # Build and deployment scripts
│   ├── build-backend.sh
│   ├── build-frontend.sh
│   ├── deploy-lambda.sh
│   ├── deploy-ecs.sh
│   └── local-dev.sh
├── cmd/                            # Go application entry points
│   ├── lambda/
│   │   └── main.go                 # Lambda entry point
│   ├── server/
│   │   └── main.go                 # HTTP server entry point
│   └── migrate/
│       └── main.go                 # Database migration tool
├── pkg/                            # Go application code (functional)
│   ├── domain/                     # Pure business logic
│   │   ├── user/
│   │   │   ├── types.go           # Immutable user types
│   │   │   ├── validation.go      # Pure validation functions
│   │   │   └── transforms.go      # Pure transformation functions
│   │   ├── message/
│   │   │   ├── types.go           # Message domain types
│   │   │   ├── validation.go      # Message validation
│   │   │   ├── business.go        # Business logic
│   │   │   └── scheduling.go      # Scheduling logic
│   │   └── common/
│   │       ├── result.go          # Result/Either monad
│   │       ├── option.go          # Option/Maybe monad
│   │       └── functional.go      # Utility functions
│   ├── handlers/                   # HTTP layer (thin)
│   │   ├── auth.go
│   │   ├── messages.go
│   │   ├── users.go
│   │   ├── health.go
│   │   └── middleware.go
│   ├── effects/                    # Side effects boundary
│   │   ├── database.go
│   │   ├── storage.go
│   │   ├── email.go
│   │   └── interfaces.go
│   ├── adapters/                   # External service adapters
│   │   ├── supabase.go
│   │   ├── s3.go
│   │   ├── ses.go
│   │   └── auth.go
│   ├── composition/                # Dependency injection
│   │   └── app.go
│   └── config/
│       ├── config.go
│       └── environment.go
├── internal/                       # Private application code
│   ├── database/
│   │   ├── migrations/
│   │   └── connection.go
│   └── utils/
│       └── helpers.go
├── ui/                             # React frontend application
│   ├── package.json
│   ├── package-lock.json
│   ├── tsconfig.json
│   ├── vite.config.ts             # Vite configuration
│   ├── tailwind.config.js         # Tailwind CSS config
│   ├── .env.example
│   ├── .env.local
│   ├── public/
│   │   ├── index.html
│   │   ├── favicon.ico
│   │   └── manifest.json
│   ├── src/
│   │   ├── main.tsx               # Application entry point
│   │   ├── App.tsx                # Root component
│   │   ├── components/            # Reusable UI components
│   │   │   ├── ui/                # Base UI components
│   │   │   │   ├── Button.tsx
│   │   │   │   ├── Input.tsx
│   │   │   │   ├── Modal.tsx
│   │   │   │   └── index.ts
│   │   │   ├── layout/            # Layout components
│   │   │   │   ├── Header.tsx
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── Layout.tsx
│   │   │   ├── auth/              # Authentication components
│   │   │   │   ├── LoginForm.tsx
│   │   │   │   ├── SignupForm.tsx
│   │   │   │   └── AuthGuard.tsx
│   │   │   └── messages/          # Message-related components
│   │   │       ├── MessageForm.tsx
│   │   │       ├── MessageList.tsx
│   │   │       ├── MessageCard.tsx
│   │   │       └── AttachmentUpload.tsx
│   │   ├── pages/                 # Page components
│   │   │   ├── Home.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   ├── CreateMessage.tsx
│   │   │   ├── MessageHistory.tsx
│   │   │   └── Profile.tsx
│   │   ├── hooks/                 # Custom React hooks
│   │   │   ├── useAuth.ts
│   │   │   ├── useMessages.ts
│   │   │   ├── useApi.ts
│   │   │   └── useLocalStorage.ts
│   │   ├── services/              # API and external services
│   │   │   ├── api.ts             # API client configuration
│   │   │   ├── auth.ts            # Authentication service
│   │   │   ├── messages.ts        # Message service
│   │   │   ├── upload.ts          # File upload service
│   │   │   └── types.ts           # TypeScript type definitions
│   │   ├── store/                 # State management
│   │   │   ├── authStore.ts       # Authentication store
│   │   │   ├── messageStore.ts    # Message store
│   │   │   └── index.ts
│   │   ├── utils/                 # Utility functions
│   │   │   ├── constants.ts
│   │   │   ├── helpers.ts
│   │   │   ├── validation.ts
│   │   │   └── formatters.ts
│   │   ├── styles/                # Global styles
│   │   │   ├── globals.css
│   │   │   └── components.css
│   │   └── types/                 # TypeScript definitions
│   │       ├── auth.ts
│   │       ├── message.ts
│   │       ├── api.ts
│   │       └── index.ts
│   ├── tests/                     # Frontend tests
│   │   ├── __mocks__/
│   │   ├── components/
│   │   ├── services/
│   │   └── utils/
│   └── dist/                      # Build output (gitignored)
├── deployments/                   # Deployment configurations
│   ├── lambda/
│   │   ├── template.yaml          # SAM template
│   │   └── Dockerfile
│   ├── ecs/
│   │   ├── docker-compose.yml
│   │   ├── task-definition.json
│   │   └── Dockerfile
│   └── k8s/
│       ├── namespace.yaml
│       ├── deployment.yaml
│       ├── service.yaml
│       └── ingress.yaml
├── tests/                         # Backend tests
│   ├── unit/
│   ├── integration/
│   └── e2e/
└── docker-compose.yml             # Local development environment
```

## Frontend Technology Stack

### Core Technologies
- **React 18** with TypeScript
- **Vite** for fast development and building
- **Tailwind CSS** for styling
- **Zustand** for state management (lightweight alternative to Redux)
- **React Query** for server state management
- **React Router** for client-side routing

### UI Components
- **Headless UI** for accessible components
- **Heroicons** for consistent iconography
- **React Hook Form** for form handling
- **Zod** for schema validation

### Development Tools
- **ESLint + Prettier** for code quality
- **Husky** for git hooks
- **Vitest** for testing
- **Testing Library** for component testing

## Frontend Package.json Structure

```json
{
  "name": "dear-future-ui",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "lint:fix": "eslint . --ext ts,tsx --fix",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.8.0",
    "axios": "^1.3.0",
    "zustand": "^4.3.0",
    "@tanstack/react-query": "^4.24.0",
    "react-hook-form": "^7.43.0",
    "zod": "^3.20.0",
    "@hookform/resolvers": "^2.9.0",
    "@headlessui/react": "^1.7.0",
    "@heroicons/react": "^2.0.0",
    "date-fns": "^2.29.0",
    "clsx": "^1.2.0",
    "tailwind-merge": "^1.10.0"
  },
  "devDependencies": {
    "@types/react": "^18.0.0",
    "@types/react-dom": "^18.0.0",
    "@vitejs/plugin-react": "^3.1.0",
    "vite": "^4.1.0",
    "typescript": "^4.9.0",
    "tailwindcss": "^3.2.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "eslint": "^8.35.0",
    "@typescript-eslint/eslint-plugin": "^5.54.0",
    "@typescript-eslint/parser": "^5.54.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "eslint-plugin-react-refresh": "^0.3.4",
    "prettier": "^2.8.0",
    "prettier-plugin-tailwindcss": "^0.2.0",
    "vitest": "^0.28.0",
    "@testing-library/react": "^14.0.0",
    "@testing-library/jest-dom": "^5.16.0",
    "@testing-library/user-event": "^14.4.0",
    "jsdom": "^21.1.0"
  }
}
```

## Development Workflow

### Local Development Setup
```bash
# Backend development
make dev-backend

# Frontend development (in parallel)
cd ui && npm run dev

# Full stack development
make dev-fullstack  # Starts both backend and frontend
```

### Build Process
```bash
# Build backend for Lambda
make build-lambda

# Build frontend for production
cd ui && npm run build

# Build everything
make build-all
```

### Testing
```bash
# Backend tests
go test ./...

# Frontend tests
cd ui && npm test

# End-to-end tests
make test-e2e
```

## Deployment Strategy

### Frontend Deployment Options

1. **Vercel (Recommended for start)**
   - Free tier with excellent performance
   - Automatic deployments from GitHub
   - Edge functions for dynamic content

2. **Netlify**
   - Free tier alternative
   - Great for static sites
   - Form handling capabilities

3. **AWS S3 + CloudFront**
   - More control over infrastructure
   - Cost-effective for higher traffic
   - Integration with AWS backend

### CI/CD Pipeline

```yaml
# .github/workflows/frontend.yml
name: Frontend CI/CD

on:
  push:
    paths: ['ui/**']
    branches: [main, staging]
  pull_request:
    paths: ['ui/**']

jobs:
  test:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ./ui
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: ui/package-lock.json
      - run: npm ci
      - run: npm run type-check
      - run: npm run lint
      - run: npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ./ui
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: ui/package-lock.json
      - run: npm ci
      - run: npm run build
      - uses: actions/upload-artifact@v3
        with:
          name: dist
          path: ui/dist

  deploy:
    if: github.ref == 'refs/heads/main'
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v3
        with:
          name: dist
          path: dist
      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.ORG_ID }}
          vercel-project-id: ${{ secrets.PROJECT_ID }}
          working-directory: ./dist
```

## Monorepo Benefits

### Development Benefits
- **Shared tooling**: Single repository for builds, linting, testing
- **Consistent dependencies**: Frontend and backend versions stay in sync
- **Atomic commits**: Changes to API and UI can be committed together
- **Simplified deployment**: Single CI/CD pipeline coordinates both components

### Code Organization
- **Clear separation**: `/ui` for frontend, root for backend
- **Shared types**: TypeScript types can be generated from Go structs
- **Common tooling**: Docker, Make, and scripts work for both components
- **Version control**: Single source of truth for the entire application

This monorepo structure provides excellent developer experience while maintaining clear separation between frontend and backend concerns, supporting the migration-ready architecture across all deployment platforms.