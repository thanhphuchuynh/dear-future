# Functional Programming Guide for Dear Future

## Overview

This guide outlines the functional programming patterns and practices used in the Dear Future codebase. Following these patterns ensures code maintainability, testability, and scalability.

## Core Functional Principles

### 1. Pure Functions
Functions that:
- Always return the same output for the same input
- Have no side effects (no I/O, no mutations)
- Don't depend on external state

```go
// ✅ Pure function
func CalculateDeliveryDate(requestedTime time.Time, timezone string) Result[time.Time] {
    loc, err := time.LoadLocation(timezone)
    if err != nil {
        return Err[time.Time](err)
    }
    return Ok(requestedTime.In(loc))
}

// ❌ Impure function (depends on external state)
func GetCurrentUser() User {
    return globalCurrentUser // depends on global state
}
```

### 2. Immutable Data Structures
Once created, data cannot be modified. Changes create new instances.

```go
// Immutable User type
type User struct {
    id    uuid.UUID
    email string
    name  string
}

// Constructor (only way to create)
func NewUser(email, name string) Result[User] {
    if err := validateEmail(email); err != nil {
        return Err[User](err)
    }
    return Ok(User{
        id:    uuid.New(),
        email: email,
        name:  name,
    })
}

// Transformation returns new instance
func (u User) WithName(newName string) Result[User] {
    return NewUser(u.email, newName)
}
```

### 3. Function Composition
Building complex operations from simple functions.

```go
// Compose validation functions
func ValidateCreateMessageRequest(req CreateMessageRequest) Result[CreateMessageRequest] {
    return pipe(
        validateContent(req.Content),
        func(string) Result[time.Time] { return validateDeliveryTime(req.DeliveryTime) },
        func(time.Time) Result[CreateMessageRequest] { return Ok(req) },
    )
}

// Pipeline helper
func pipe[T, U any](input Result[T], fn func(T) Result[U]) Result[U] {
    if input.IsErr() {
        return Err[U](input.Error())
    }
    return fn(input.Value())
}
```

## Error Handling with Result Type

### Result Monad Implementation
```go
type Result[T any] struct {
    value T
    err   error
}

func Ok[T any](value T) Result[T] {
    return Result[T]{value: value}
}

func Err[T any](err error) Result[T] {
    return Result[T]{err: err}
}

func (r Result[T]) IsOk() bool { return r.err == nil }
func (r Result[T]) IsErr() bool { return r.err != nil }
func (r Result[T]) Value() T { return r.value }
func (r Result[T]) Error() error { return r.err }
```

### Using Result for Error Handling
```go
// Chain operations that might fail
func ProcessMessage(req CreateMessageRequest) Result[Message] {
    return pipe(
        validateRequest(req),
        transformToMessage,
        enrichWithMetadata,
        validateBusinessRules,
    )
}

// Handle results
func (h *MessageHandler) CreateMessage(c *gin.Context) {
    result := ProcessMessage(parseRequest(c))
    
    if result.IsErr() {
        c.JSON(400, ErrorResponse{Error: result.Error().Error()})
        return
    }
    
    // Save to database (side effect)
    saveResult := h.db.SaveMessage(c, result.Value())
    respondWithResult(c, saveResult)
}
```

## Side Effect Management

### Isolating Side Effects
All I/O operations are isolated to specific interfaces:

```go
// Effects are isolated in interfaces
type Database interface {
    SaveMessage(ctx context.Context, msg Message) Result[Message]
    FindMessages(ctx context.Context, userID string) Result[[]Message]
}

type EmailService interface {
    SendEmail(ctx context.Context, msg Message, user User) Result[SendResult]
}

// Business logic is pure
func ScheduleMessage(msg Message, user User) ScheduleResult {
    // Pure logic only - no side effects
    deliveryTime := calculateOptimalDeliveryTime(msg.RequestedTime, user.Timezone)
    return ScheduleResult{
        Message:      msg,
        DeliveryTime: deliveryTime,
        Recipient:    user,
    }
}
```

### Dependency Injection Pattern
```go
type MessageService struct {
    db    Database
    email EmailService
}

func NewMessageService(db Database, email EmailService) *MessageService {
    return &MessageService{db: db, email: email}
}

func (s *MessageService) CreateMessage(ctx context.Context, req CreateMessageRequest) Result[Message] {
    // Pure business logic
    messageResult := ProcessMessage(req)
    if messageResult.IsErr() {
        return messageResult
    }
    
    // Side effect: save to database
    return s.db.SaveMessage(ctx, messageResult.Value())
}
```

## Testing Functional Code

### Testing Pure Functions
```go
func TestCalculateDeliveryDate(t *testing.T) {
    tests := []struct {
        name         string
        requestedTime time.Time
        timezone     string
        expectErr    bool
    }{
        {
            name:         "valid UTC time",
            requestedTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
            timezone:     "UTC",
            expectErr:    false,
        },
        {
            name:         "invalid timezone",
            requestedTime: time.Now(),
            timezone:     "Invalid/Timezone",
            expectErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateDeliveryDate(tt.requestedTime, tt.timezone)
            
            if tt.expectErr {
                assert.True(t, result.IsErr())
            } else {
                assert.True(t, result.IsOk())
            }
        })
    }
}
```

### Testing with Mock Dependencies
```go
type MockDatabase struct {
    messages map[string]Message
}

func (m *MockDatabase) SaveMessage(ctx context.Context, msg Message) Result[Message] {
    m.messages[msg.ID] = msg
    return Ok(msg)
}

func TestMessageService_CreateMessage(t *testing.T) {
    db := &MockDatabase{messages: make(map[string]Message)}
    service := NewMessageService(db, &MockEmailService{})
    
    req := CreateMessageRequest{
        Content:      "Test message",
        DeliveryTime: time.Now().Add(24 * time.Hour),
    }
    
    result := service.CreateMessage(context.Background(), req)
    
    assert.True(t, result.IsOk())
    assert.Len(t, db.messages, 1)
}
```

## Performance Considerations

### Avoiding Unnecessary Allocations
```go
// Use value receivers for immutable types
func (u User) GetDisplayName() string {
    if u.name != "" {
        return u.name
    }
    return u.email
}

// Use pointer receivers only for methods that need them
func (s *MessageService) processLargeBatch(messages []Message) Result[[]ProcessedMessage] {
    // Process large data efficiently
}
```

### Memory-Efficient Function Composition
```go
// Use interfaces for function composition
type Validator[T any] func(T) Result[T]
type Transformer[T, U any] func(T) Result[U]

// Compose without creating intermediate slices
func ComposeValidators[T any](validators ...Validator[T]) Validator[T] {
    return func(input T) Result[T] {
        current := Ok(input)
        for _, validator := range validators {
            if current.IsErr() {
                return current
            }
            current = validator(current.Value())
        }
        return current
    }
}
```

## Common Patterns

### Option/Maybe Type for Nullable Values
```go
type Option[T any] struct {
    value T
    hasValue bool
}

func Some[T any](value T) Option[T] {
    return Option[T]{value: value, hasValue: true}
}

func None[T any]() Option[T] {
    return Option[T]{hasValue: false}
}

func (o Option[T]) IsSome() bool { return o.hasValue }
func (o Option[T]) IsNone() bool { return !o.hasValue }
func (o Option[T]) Value() T { return o.value }

// Usage
func FindUserByEmail(email string) Option[User] {
    // Return Some(user) if found, None() if not found
}
```

### Higher-Order Functions
```go
// Map function for slices
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, item := range slice {
        result[i] = fn(item)
    }
    return result
}

// Filter function
func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, item := range slice {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

// Reduce function
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
    result := initial
    for _, item := range slice {
        result = fn(result, item)
    }
    return result
}
```

## Best Practices

### Do's
- ✅ Write pure functions for business logic
- ✅ Use immutable data structures
- ✅ Isolate side effects in interfaces
- ✅ Compose functions to build complex operations
- ✅ Use Result/Option types for error handling
- ✅ Write comprehensive tests for pure functions

### Don'ts
- ❌ Mix business logic with I/O operations
- ❌ Use global variables or mutable shared state
- ❌ Create functions with side effects in domain layer
- ❌ Ignore error handling with proper types
- ❌ Make data structures mutable without good reason

## Migration Notes

When migrating between deployment platforms (Lambda → ECS → K8s), the functional architecture ensures:

1. **Business logic remains unchanged** - pure functions work everywhere
2. **Only I/O adapters change** - database/email implementations
3. **Tests remain valid** - pure functions test the same way
4. **Configuration is externalized** - no hardcoded dependencies

This makes platform migration significantly easier and less risky.