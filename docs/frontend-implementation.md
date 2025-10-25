# Frontend Implementation Guide

## Overview

The Dear Future frontend is built with **Next.js 15**, **React 19**, **TypeScript**, and **Tailwind CSS 4**. It provides a modern, responsive interface for users to create and manage messages to their future selves.

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **UI Library**: React 19
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4
- **Build Tool**: Turbopack (Next.js default)
- **State Management**: React Context API

## Project Structure

```
ui/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── auth/
│   │   │   ├── login/         # Login page
│   │   │   └── register/      # Registration page
│   │   ├── dashboard/         # Dashboard page (protected)
│   │   ├── messages/          # Messages pages
│   │   │   ├── [id]/         # Edit message (dynamic route)
│   │   │   ├── new/          # Create new message
│   │   │   └── page.tsx      # Messages list
│   │   ├── layout.tsx         # Root layout with AuthProvider
│   │   ├── page.tsx           # Landing page
│   │   └── globals.css        # Global styles
│   ├── components/            # Reusable components
│   │   ├── Navigation.tsx     # Main navigation bar
│   │   └── ProtectedRoute.tsx # Route protection wrapper
│   ├── contexts/              # React contexts
│   │   └── AuthContext.tsx    # Authentication state management
│   └── lib/                   # Utilities and services
│       ├── api-client.ts      # API client for backend communication
│       └── types.ts           # TypeScript type definitions
├── .env.local                 # Environment variables
├── next.config.ts             # Next.js configuration
├── tailwind.config.ts         # Tailwind CSS configuration
├── tsconfig.json              # TypeScript configuration
└── package.json               # Dependencies
```

## Key Features

### 1. Authentication System

**Location**: `src/contexts/AuthContext.tsx`, `src/app/auth/`

- JWT-based authentication
- Persistent login state (localStorage)
- Protected routes with automatic redirect
- Login and registration pages with validation

**Usage**:
```tsx
import { useAuth } from '@/contexts/AuthContext';

function MyComponent() {
  const { user, isAuthenticated, login, logout } = useAuth();
  // ...
}
```

### 2. API Client

**Location**: `src/lib/api-client.ts`

- Singleton pattern for consistent token management
- Automatic token injection for authenticated requests
- Type-safe API calls with TypeScript
- Error handling and response parsing

**Example**:
```typescript
import { apiClient } from '@/lib/api-client';

// Register user
const response = await apiClient.register({
  email: 'user@example.com',
  password: 'SecurePass123',
  name: 'John Doe'
});

// Create message
const message = await apiClient.createMessage({
  title: 'Future Goals',
  content: 'Remember to...',
  delivery_date: '2026-01-01T00:00:00Z',
  timezone: 'UTC',
  delivery_method: 'email'
});
```

### 3. Protected Routes

**Location**: `src/components/ProtectedRoute.tsx`

Wraps authenticated pages and redirects to login if not authenticated:

```tsx
<ProtectedRoute>
  <YourProtectedContent />
</ProtectedRoute>
```

### 4. Pages

#### Landing Page (`/`)
- Hero section with call-to-action
- Feature highlights
- Automatic redirect to dashboard if authenticated

#### Login Page (`/auth/login`)
- Email/password authentication
- Link to registration
- Error handling

#### Registration Page (`/auth/register`)
- User registration form
- Password validation (8+ chars, letters & numbers)
- Automatic timezone detection

#### Dashboard (`/dashboard`)
- Statistics cards (total, scheduled, sent messages)
- Quick actions
- Recent messages preview
- Protected route

#### Messages List (`/messages`)
- Filter by status (all, scheduled, sent)
- Message cards with status badges
- Edit and delete actions
- Protected route

#### New Message (`/messages/new`)
- Create message form
- DateTime picker for delivery
- Delivery method selector
- Protected route

#### Edit Message (`/messages/[id]`)
- Edit existing message
- View message metadata
- Delete message option
- Protected route

## API Integration

The frontend communicates with the backend through the API client:

### Authentication Endpoints
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### User Endpoints
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile

### Message Endpoints
- `GET /api/v1/messages` - List all messages
- `POST /api/v1/messages` - Create new message
- `GET /api/v1/messages/:id` - Get message by ID
- `PUT /api/v1/messages/:id` - Update message
- `DELETE /api/v1/messages/:id` - Delete message

## Environment Variables

Create `.env.local` in the `ui/` directory:

```bash
# Backend API URL (used by Next.js rewrites)
NEXT_PUBLIC_API_URL=http://localhost:8080

# Environment
NEXT_PUBLIC_ENVIRONMENT=development
```

**Note**: `NEXT_PUBLIC_` prefix makes variables available in the browser.

## Next.js Configuration

**Location**: `ui/next.config.ts`

Key configuration:
- **API Rewrites**: Proxies `/api/*` requests to backend on `localhost:8080`
- This avoids CORS issues in development

```typescript
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8080/api/:path*',
    }
  ];
}
```

## Styling

### Tailwind CSS

The project uses **Tailwind CSS 4** with:
- Dark mode support (via `dark:` prefix)
- Responsive design (via `sm:`, `md:`, `lg:` breakpoints)
- Custom color scheme (blue/indigo theme)

### Component Styling Pattern

```tsx
<button className="
  px-4 py-2                           // Padding
  bg-blue-600 hover:bg-blue-700       // Background colors
  text-white                          // Text color
  rounded-lg                          // Border radius
  font-medium                         // Font weight
  transition-colors                   // Smooth transitions
  disabled:opacity-50                 // Disabled state
  dark:bg-blue-500                    // Dark mode
">
  Button Text
</button>
```

## Development Workflow

### Start Development Server

```bash
# From project root
make dev               # Start all services (DB + Backend + UI)

# Or individually
cd ui && npm run dev   # Start only UI on http://localhost:3000
```

### Build for Production

```bash
cd ui && npm run build
cd ui && npm start      # Production server
```

### Type Checking

```bash
cd ui && npx tsc --noEmit
```

### Linting

```bash
cd ui && npm run lint
```

## State Management

### Authentication State

Managed by `AuthContext`:
- User object
- Loading state
- Authentication status
- Login/logout functions

### Local State

Individual components use React hooks:
- `useState` for form data and UI state
- `useEffect` for data fetching
- `useRouter` for navigation

## Data Flow

```
User Action
    ↓
Component Handler
    ↓
API Client (with JWT token)
    ↓
Next.js Rewrite (/api → http://localhost:8080/api)
    ↓
Backend Server
    ↓
Response
    ↓
API Client (parse JSON)
    ↓
Component Update (setState)
    ↓
UI Re-render
```

## Authentication Flow

```
1. User visits site
   ↓
2. AuthContext checks localStorage for token
   ↓
3. If token exists → Validate with backend
   ↓
4. If valid → Set user state, allow access
   ↓
5. If invalid → Clear token, redirect to login
```

## Protected Route Flow

```
User navigates to /dashboard
    ↓
ProtectedRoute component renders
    ↓
Check isAuthenticated from AuthContext
    ↓
If TRUE → Render children (Dashboard)
    ↓
If FALSE → Redirect to /auth/login
```

## TypeScript Types

All types are defined in `src/lib/types.ts`:

```typescript
// Example types
export interface User {
  id: string;
  email: string;
  name?: string;
  timezone?: string;
  created_at: string;
}

export interface Message {
  id: string;
  user_id: string;
  title: string;
  content: string;
  delivery_date: string;
  timezone: string;
  status: 'scheduled' | 'sent' | 'failed' | 'cancelled';
  delivery_method: 'email' | 'sms' | 'push';
  created_at: string;
  updated_at: string;
}
```

## Common Tasks

### Add a New Page

1. Create file in `src/app/new-page/page.tsx`
2. If protected, wrap with `<ProtectedRoute>`
3. Add navigation link in `Navigation.tsx`

### Add New API Endpoint

1. Add method to `api-client.ts`
2. Add types to `types.ts` if needed
3. Use in component

### Customize Styling

1. Edit Tailwind classes in components
2. Or add custom CSS in `globals.css`
3. Update `tailwind.config.ts` for theme changes

## Best Practices

1. **Always use TypeScript types** - Import from `@/lib/types`
2. **Handle loading states** - Show spinners during API calls
3. **Handle errors gracefully** - Display user-friendly error messages
4. **Use environment variables** - Never hardcode URLs
5. **Protected routes** - Wrap authenticated pages with `ProtectedRoute`
6. **Responsive design** - Test on mobile, tablet, desktop
7. **Dark mode** - Add `dark:` variants for all colors

## Troubleshooting

### Port Already in Use

If port 3000 is occupied:
```bash
PORT=3001 npm run dev
```

### API Requests Failing

- Check backend is running on port 8080
- Check `.env.local` has correct `NEXT_PUBLIC_API_URL`
- Check browser console for CORS errors

### Type Errors

```bash
npx tsc --noEmit  # Check for type errors
```

### Build Errors

```bash
rm -rf .next      # Clear Next.js cache
npm run build     # Rebuild
```

## Performance Optimizations

1. **Next.js App Router** - Automatic code splitting
2. **Turbopack** - Fast development builds
3. **Server Components** - Reduced client-side JavaScript (not used yet, but available)
4. **Image Optimization** - Use Next.js `<Image>` component
5. **Font Optimization** - Geist fonts loaded via Next.js

## Security Considerations

1. **JWT Storage** - Tokens in localStorage (consider httpOnly cookies for production)
2. **XSS Prevention** - React escapes content automatically
3. **CSRF** - JWT tokens are not vulnerable to CSRF
4. **HTTPS** - Use in production
5. **Environment Variables** - Never commit `.env.local` to git

## Testing URLs

Once running, test these URLs:

- Landing page: http://localhost:3000
- Login: http://localhost:3000/auth/login
- Register: http://localhost:3000/auth/register
- Dashboard: http://localhost:3000/dashboard (requires login)
- Messages: http://localhost:3000/messages (requires login)
- New Message: http://localhost:3000/messages/new (requires login)

## Next Steps

To continue development:

1. **Add real PostgreSQL persistence** - Currently using mock database
2. **Implement file attachments** - Use Cloudflare R2 adapter
3. **Add email templates** - Customize delivery emails
4. **User settings page** - Profile editing, timezone selection
5. **Message scheduling** - Add recurring messages
6. **Analytics dashboard** - Track sent/opened messages
7. **Mobile app** - React Native with shared types

## Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [React Documentation](https://react.dev)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [TypeScript Documentation](https://www.typescriptlang.org/docs)
