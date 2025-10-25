# Dear Future API - Quick Start Guide

## Overview

The Dear Future API allows you to create and manage messages to your future self. This guide will help you get started quickly.

## Base URL

```
Development: http://localhost:8080
Production: https://api.dearfuture.app
```

## Quick Start (5 Minutes)

### 1. Start the Server

```bash
# Clone and setup
git clone https://github.com/your-username/dear-future.git
cd dear-future

# Start database
docker-compose up -d postgres

# Start server
make dev-backend

# Or manually
go run main.go
```

Server will start on `http://localhost:8080`

### 2. Test the API

Run the automated test script:
```bash
./scripts/test-api.sh
```

Or follow the manual steps below.

## Authentication Flow

The API uses JWT (JSON Web Tokens) for authentication.

### Step 1: Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "you@example.com",
    "name": "Your Name",
    "password": "SecurePass123",
    "timezone": "America/New_York"
  }'
```

**Response**:
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "you@example.com",
    "name": "Your Name",
    "timezone": "America/New_York",
    "created_at": "2025-10-25T12:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2025-10-25T12:15:00Z"
}
```

**Save your tokens!** You'll need the `access_token` for authenticated requests.

### Step 2: Create a Message

```bash
# Replace YOUR_TOKEN with the access_token from Step 1
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello Future Me!",
    "content": "Remember to stay positive and keep learning!",
    "delivery_date": "2026-01-01T00:00:00Z",
    "timezone": "America/New_York",
    "delivery_method": "email"
  }'
```

**Response**:
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Hello Future Me!",
  "content": "Remember to stay positive and keep learning!",
  "delivery_date": "2026-01-01T00:00:00Z",
  "timezone": "America/New_York",
  "status": "scheduled",
  "delivery_method": "email",
  "created_at": "2025-10-25T12:00:00Z",
  "updated_at": "2025-10-25T12:00:00Z"
}
```

### Step 3: View Your Messages

```bash
curl http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Common Use Cases

### Use Case 1: Birthday Message

Send yourself a birthday message for next year:

```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Happy Birthday! 🎂",
    "content": "Another year older, another year wiser! Hope you achieved your goals this year.",
    "delivery_date": "2026-10-25T09:00:00Z",
    "timezone": "America/New_York",
    "delivery_method": "email"
  }'
```

### Use Case 2: New Year Resolution Check-in

```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Year Resolution Check-in",
    "content": "Did you stick to your resolutions? Remember: progress over perfection!",
    "delivery_date": "2026-06-01T00:00:00Z",
    "timezone": "UTC",
    "delivery_method": "email"
  }'
```

### Use Case 3: Career Milestone

```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Career Check-in",
    "content": "Are you where you wanted to be career-wise? If not, what steps can you take now?",
    "delivery_date": "2027-01-01T00:00:00Z",
    "timezone": "UTC",
    "delivery_method": "email"
  }'
```

## API Endpoints Cheat Sheet

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login existing user |
| POST | `/api/v1/auth/refresh` | Refresh access token |

### User Profile

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/user/profile` | ✅ | Get user profile |
| PUT | `/api/v1/user/update` | ✅ | Update profile |

### Messages

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/messages` | ✅ | Create message |
| GET | `/api/v1/messages` | ✅ | List all messages |
| GET | `/api/v1/messages?id={id}` | ✅ | Get single message |
| PUT | `/api/v1/messages?id={id}` | ✅ | Update message |
| DELETE | `/api/v1/messages?id={id}` | ✅ | Delete message |

### System

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | ❌ | Health check |
| GET | `/api/v1/` | ❌ | API information |

## Authentication

All authenticated endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer YOUR_ACCESS_TOKEN
```

### Token Expiration

- **Access Token**: Expires after 15 minutes
- **Refresh Token**: Expires after 7 days

### Refreshing Tokens

When your access token expires, use the refresh token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

## Error Handling

All errors return this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

- **200**: Success (GET, PUT, DELETE)
- **201**: Created (POST)
- **400**: Bad Request (invalid input)
- **401**: Unauthorized (invalid/missing token)
- **403**: Forbidden (no permission)
- **404**: Not Found
- **500**: Server Error

## Password Requirements

- Minimum 8 characters
- Must contain at least one letter
- Must contain at least one number
- Maximum 72 characters

## Query Parameters

### List Messages

```bash
# Get first 10 messages
curl "http://localhost:8080/api/v1/messages?limit=10&offset=0" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get next 10 messages
curl "http://localhost:8080/api/v1/messages?limit=10&offset=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

- `limit`: Maximum messages to return (default: 50, max: 100)
- `offset`: Number of messages to skip (default: 0)

## Date Formatting

All dates use ISO 8601 / RFC3339 format:

```
YYYY-MM-DDTHH:MM:SSZ

Examples:
  2026-01-01T00:00:00Z        # UTC
  2026-01-01T00:00:00-05:00   # EST
  2026-12-31T23:59:59+00:00   # UTC
```

## Testing Tools

### cURL Examples

See examples throughout this guide.

### Automated Testing

Run the complete test suite:

```bash
./scripts/test-api.sh
```

### Postman/Thunder Client

Import the API collection:
```bash
# Collection file location
./docs/api-collection.json
```

## Environment Variables

### Required for Production

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/dear_future

# JWT Secret (change this!)
JWT_SECRET=your-super-secret-key-min-32-characters

# Optional: Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Optional: R2 Storage
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-key
R2_SECRET_ACCESS_KEY=your-secret
R2_BUCKET_NAME=dear-future-uploads
```

## Rate Limiting

Currently no rate limiting implemented. In production:
- 60 requests per minute per IP
- 100 requests per minute per user

## CORS

Development mode allows all origins. Production restricts to:
- https://dearfuture.app
- https://www.dearfuture.app

## WebSocket Support

Not yet implemented. Planned for:
- Real-time message updates
- Delivery notifications
- System alerts

## Webhooks

Not yet implemented. Planned for:
- Message delivered event
- Message failed event
- User action events

## SDKs

### Official SDKs

Not yet available. Coming soon:
- JavaScript/TypeScript
- Python
- Ruby
- Go

### Community SDKs

None yet. Contributions welcome!

## Best Practices

### 1. Store Tokens Securely

```javascript
// ❌ Don't store in localStorage
localStorage.setItem('token', accessToken);

// ✅ Use httpOnly cookies or secure storage
// Server should set httpOnly cookie
```

### 2. Handle Token Expiration

```javascript
// Implement automatic token refresh
if (response.status === 401) {
  const newToken = await refreshAccessToken();
  // Retry request with new token
}
```

### 3. Validate Input Client-Side

```javascript
// Validate before sending to API
if (password.length < 8) {
  return "Password must be at least 8 characters";
}
```

### 4. Use Environment-Specific URLs

```javascript
const API_URL = process.env.NODE_ENV === 'production'
  ? 'https://api.dearfuture.app'
  : 'http://localhost:8080';
```

## Troubleshooting

### "Invalid token" Error

- Check token hasn't expired (15 min for access tokens)
- Verify Bearer prefix in Authorization header
- Use refresh token to get new access token

### "Password must contain..." Error

- Ensure password has letters AND numbers
- Check password length (8-72 characters)

### Connection Refused

- Verify server is running (`make dev-backend`)
- Check correct port (default: 8080)
- Ensure database is running (`docker-compose up -d`)

### CORS Errors

- In development, CORS is open
- In production, check your origin is whitelisted
- Use API proxy if needed

## Support

- **Documentation**: [docs/api-implementation.md](docs/api-implementation.md)
- **Issues**: [GitHub Issues](https://github.com/your-username/dear-future/issues)
- **Discord**: [Join our community](https://discord.gg/dearfuture)

## Next Steps

1. ✅ Complete this quickstart
2. 📚 Read full API documentation
3. 🎨 Build your frontend
4. 🚀 Deploy to production
5. 🌟 Share your experience!

---

**Happy coding!** 🚀

Built with ❤️ using functional programming principles in Go.
