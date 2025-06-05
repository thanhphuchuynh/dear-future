# Configuration Guide

Dear Future uses a flexible YAML-based configuration system with environment variable overrides, supporting multiple deployment environments while maintaining functional programming principles.

## 📁 Configuration Files

### Default Configuration
- **File**: `config.yaml`
- **Environment**: Development
- **Use Case**: Local development and testing

### Staging Configuration
- **File**: `config.staging.yaml`
- **Environment**: Staging
- **Use Case**: Pre-production testing and QA

### Production Configuration
- **File**: `config.production.yaml`
- **Environment**: Production
- **Use Case**: Live deployment

## 🏗️ Configuration Architecture

### Loading Priority (Highest to Lowest)
1. **Environment Variables** (Highest priority)
2. **YAML Configuration File**
3. **Default Values** (Fallback)

### File Selection Methods
```bash
# Method 1: CLI flag
./dear-future-cli --config config.staging.yaml --cmd health

# Method 2: Environment variable
CONFIG_FILE=config.staging.yaml ./dear-future

# Method 3: Default
./dear-future  # Uses config.yaml
```

## ⚙️ Configuration Structure

### Server Configuration
```yaml
server:
  port: "8080"                # Override: PORT
  read_timeout: "15s"         # Override: READ_TIMEOUT
  write_timeout: "15s"        # Override: WRITE_TIMEOUT
  idle_timeout: "60s"         # Override: IDLE_TIMEOUT
```

### Database Configuration
```yaml
database:
  url: ""                     # Override: DATABASE_URL (Required)
  max_conns: 25              # Override: DATABASE_MAX_CONNS
  max_idle_conns: 5          # Override: DATABASE_MAX_IDLE_CONNS
  conn_lifetime: "5m"        # Override: DATABASE_CONN_LIFETIME
```

### Authentication Configuration
```yaml
auth:
  jwt_secret: ""             # Override: JWT_SECRET (Required in prod)
  jwt_expiration: "15m"      # Override: JWT_EXPIRATION
  refresh_token_lifetime: "168h"  # Override: REFRESH_TOKEN_LIFETIME
  password_min_length: 8     # Override: PASSWORD_MIN_LENGTH
```

### AWS Services Configuration
```yaml
aws:
  region: "us-east-1"        # Override: AWS_REGION
  s3_bucket: ""              # Override: S3_BUCKET
  s3_bucket_region: "us-east-1"  # Override: S3_BUCKET_REGION
  ses_region: "us-east-1"    # Override: SES_REGION
  ses_from_email: "noreply@dearfuture.app"  # Override: SES_FROM_EMAIL
  ses_from_name: "Dear Future"  # Override: SES_FROM_NAME
```

### Supabase Configuration
```yaml
supabase:
  url: ""                    # Override: SUPABASE_URL
  anon_key: ""              # Override: SUPABASE_ANON_KEY
  service_key: ""           # Override: SUPABASE_SERVICE_KEY
```

### Feature Flags
```yaml
features:
  push_notifications: false  # Override: FEATURE_PUSH_NOTIFICATIONS
  webhooks: false           # Override: FEATURE_WEBHOOKS
  message_templates: false  # Override: FEATURE_MESSAGE_TEMPLATES
  batch_processing: true    # Override: FEATURE_BATCH_PROCESSING
  advanced_scheduling: false # Override: FEATURE_ADVANCED_SCHEDULING
  file_attachments: false   # Override: FEATURE_FILE_ATTACHMENTS
  email_reminders: false    # Override: FEATURE_EMAIL_REMINDERS
  analytics: false          # Override: FEATURE_ANALYTICS
```

## 🌍 Environment-Specific Configurations

### Development Environment (`config.yaml`)
```yaml
environment: development
debug: true
log_level: info

features:
  file_attachments: false    # Disabled by default
  analytics: false
  
security:
  cors_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  rate_limit_per_minute: 60
```

### Staging Environment (`config.staging.yaml`)
```yaml
environment: staging
debug: true
log_level: info

features:
  file_attachments: true     # Enabled for testing
  analytics: false          # Disabled in staging
  
file_upload:
  max_file_size: 52428800   # 50MB

security:
  cors_origins:
    - "https://staging.dearfuture.app"
  rate_limit_per_minute: 100
```

### Production Environment (`config.production.yaml`)
```yaml
environment: production
debug: false
log_level: warn

features:
  file_attachments: true     # Fully enabled
  analytics: true           # Enabled for insights
  
file_upload:
  max_file_size: 104857600  # 100MB

auth:
  password_min_length: 12   # Stricter security

security:
  cors_origins:
    - "https://dearfuture.app"
    - "https://www.dearfuture.app"
  rate_limit_per_minute: 120
  enable_https: true
```

## 🚀 Usage Examples

### Development Setup
```bash
# Default development
make run-dev

# With custom port
PORT=9000 make run-dev

# With verbose logging
DEBUG=true LOG_LEVEL=debug make run-dev
```

### Staging Deployment
```bash
# Using Makefile
make run-staging

# Manual with environment variables
CONFIG_FILE=config.staging.yaml \
DATABASE_URL=postgresql://staging:password@db.staging.com/dearfuture \
S3_BUCKET=dear-future-staging \
JWT_SECRET=staging-secret \
./bin/dear-future
```

### Production Deployment
```bash
# Using Makefile
make run-prod

# Manual with environment variables
CONFIG_FILE=config.production.yaml \
DATABASE_URL=postgresql://prod:secret@db.prod.com/dearfuture \
S3_BUCKET=dear-future-production \
JWT_SECRET=super-secure-production-secret \
SUPABASE_URL=https://your-project.supabase.co \
SUPABASE_SERVICE_KEY=your-service-key \
./bin/dear-future
```

### CLI Health Checks
```bash
# Development
make cli-health

# Staging
make cli-health-staging

# Production  
make cli-health-prod

# Custom config
./bin/dear-future-cli --config my-config.yaml --cmd health
```

## 🔒 Security Considerations

### Required Environment Variables

#### Development
- `DATABASE_URL` or `SUPABASE_URL`

#### Staging
- `DATABASE_URL` or `SUPABASE_URL`
- `S3_BUCKET` (if file attachments enabled)
- `JWT_SECRET`

#### Production
- `DATABASE_URL` or `SUPABASE_URL`
- `S3_BUCKET`
- `JWT_SECRET` (Must not use default)
- `SUPABASE_SERVICE_KEY`

### Security Validations
- JWT secret cannot be default value in production
- S3 bucket required when file attachments enabled
- HTTPS enforced in production configuration
- Rate limiting automatically configured per environment

## 🛠️ Configuration Management

### Creating Custom Configurations
```yaml
# config.custom.yaml
environment: custom
debug: false

# Override specific sections
server:
  port: "3000"
  
features:
  file_attachments: true
  analytics: true

# Use with:
CONFIG_FILE=config.custom.yaml ./bin/dear-future
```

### Environment Variable Patterns
```bash
# Boolean values
DEBUG=true
FEATURE_FILE_ATTACHMENTS=false

# Numeric values
PORT=8080
DATABASE_MAX_CONNS=50
MAX_FILE_SIZE=104857600

# Duration values
READ_TIMEOUT=30s
JWT_EXPIRATION=1h
CACHE_TTL=10m

# Array values (comma-separated)
CORS_ORIGINS="https://app.com,https://www.app.com"
ALLOWED_FILE_TYPES="image/jpeg,image/png,application/pdf"
```

### Validation Rules
- Environment must be specified
- Database URL or Supabase URL required
- File attachments require S3 bucket configuration
- Production requires secure JWT secret
- Timeouts must be positive durations
- File size limits must be positive

## 📊 Configuration Testing

### Validate Configuration
```bash
# Test configuration loading
./bin/dear-future-cli --config config.staging.yaml --cmd version

# Health check with configuration
./bin/dear-future-cli --config config.production.yaml --cmd health

# Override specific values
PORT=9999 ./bin/dear-future-cli --cmd health
```

### Common Issues and Solutions

#### Issue: S3_BUCKET required error
```bash
# Solution: Disable file attachments or provide bucket
FEATURE_FILE_ATTACHMENTS=false ./bin/dear-future
# OR
S3_BUCKET=my-bucket ./bin/dear-future
```

#### Issue: JWT secret error in production
```bash
# Solution: Set secure JWT secret
JWT_SECRET=your-secure-secret CONFIG_FILE=config.production.yaml ./bin/dear-future
```

#### Issue: Database connection error
```bash
# Solution: Provide valid database URL
DATABASE_URL=postgresql://user:pass@host:5432/db ./bin/dear-future
```

## 🎯 Best Practices

1. **Never commit secrets**: Use environment variables for sensitive data
2. **Environment-specific configs**: Use separate files for each environment
3. **Validate early**: Test configuration before deployment
4. **Document overrides**: Clearly document which environment variables override YAML
5. **Use feature flags**: Enable/disable features per environment
6. **Security first**: Enforce stricter settings in production
7. **Monitor configuration**: Log configuration loading and validation

## 🔗 Integration with Deployment

### AWS Lambda
```bash
# Set environment variables in Lambda configuration
CONFIG_FILE=config.production.yaml
DATABASE_URL=postgresql://...
S3_BUCKET=prod-bucket
JWT_SECRET=lambda-secret
```

### Docker
```dockerfile
# Copy config files
COPY config.*.yaml ./

# Set environment variables
ENV CONFIG_FILE=config.production.yaml
ENV DATABASE_URL=postgresql://...
```

### Kubernetes
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dear-future-config
data:
  config.yaml: |
    environment: production
    # ... rest of config

---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: dear-future
        env:
        - name: CONFIG_FILE
          value: "/config/config.yaml"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
```

This configuration system provides maximum flexibility while maintaining the functional programming principles of immutability and pure configuration loading.