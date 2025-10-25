# Infrastructure Setup Guide

This guide explains how to set up the infrastructure services for Dear Future application in production.

## Overview

The application uses:
- **PostgreSQL** for database (local dev or Supabase for production)
- **Cloudflare R2** for file storage (S3-compatible, cheaper than AWS S3)
- **Gmail SMTP** for sending emails (free for moderate use)

## Table of Contents

1. [Database Setup](#database-setup)
2. [Cloudflare R2 Storage Setup](#cloudflare-r2-storage-setup)
3. [Gmail SMTP Setup](#gmail-smtp-setup)
4. [Environment Configuration](#environment-configuration)
5. [Testing Your Setup](#testing-your-setup)

---

## Database Setup

### Local Development (Docker PostgreSQL)

For local development, use the included Docker Compose setup:

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Database will be available at:
# postgresql://postgres:postgres@localhost:5432/dear_future_dev?sslmode=disable
```

### Production (Supabase)

1. **Create a Supabase Project**
   - Go to [https://supabase.com](https://supabase.com)
   - Click "New Project"
   - Choose organization, name, and region
   - Set a strong database password

2. **Get Connection Details**
   - Go to Project Settings > Database
   - Copy the connection string
   - Go to Project Settings > API
   - Copy the project URL and service role key

3. **Run Database Migrations**
   ```bash
   # Use the SQL Editor in Supabase dashboard
   # Or run migrations via psql
   psql "your-connection-string" < migrations/001_init_schema.sql
   ```

4. **Configure Environment Variables**
   ```bash
   DATABASE_URL=postgresql://postgres:[password]@db.[project].supabase.co:5432/postgres
   SUPABASE_URL=https://[project].supabase.co
   SUPABASE_SERVICE_KEY=your-service-role-key
   ```

---

## Cloudflare R2 Storage Setup

Cloudflare R2 is S3-compatible object storage with **zero egress fees**, making it significantly cheaper than AWS S3 for public-facing applications.

### Why R2 over AWS S3?

- **No egress fees**: Data transfer out is free
- **S3-compatible**: Works with existing S3 SDKs
- **Cheaper storage**: $0.015/GB/month vs S3's $0.023/GB/month
- **Better for user-facing content**: No surprise bills from bandwidth usage

### Setup Steps

1. **Create Cloudflare Account**
   - Go to [https://dash.cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up)
   - Complete email verification

2. **Enable R2**
   - Go to R2 section in dashboard
   - Click "Purchase R2 Plan" (includes 10GB free storage + 1M Class A operations per month)

3. **Create an R2 Bucket**
   ```bash
   # Go to R2 > Create bucket
   # Or use Wrangler CLI:
   npx wrangler r2 bucket create dear-future-uploads
   ```

   Settings:
   - **Bucket name**: `dear-future-uploads` (or your choice)
   - **Location**: Automatic (Cloudflare will optimize)

4. **Create API Token**
   - Go to R2 > Manage R2 API Tokens
   - Click "Create API Token"
   - Permissions: "Object Read & Write"
   - Select your bucket or "All buckets"
   - Copy the **Access Key ID** and **Secret Access Key**

5. **Get Your Account ID**
   - Found in the URL: `https://dash.cloudflare.com/[ACCOUNT_ID]/r2`
   - Or in Account settings

6. **Configure Public Access (Optional)**

   If you want files to be publicly accessible:

   ```bash
   # Option 1: Custom Domain (Recommended)
   # Go to Bucket Settings > Public Access
   # Add custom domain: uploads.yourdomain.com

   # Option 2: R2.dev Subdomain
   # Enable R2.dev subdomain in bucket settings
   # URL format: https://[bucket-name].[account-id].r2.dev
   ```

7. **Environment Variables**
   ```bash
   R2_ACCOUNT_ID=your-cloudflare-account-id
   R2_ACCESS_KEY_ID=your-access-key-id
   R2_SECRET_ACCESS_KEY=your-secret-access-key
   R2_BUCKET_NAME=dear-future-uploads
   R2_PUBLIC_URL=https://uploads.yourdomain.com  # optional
   ```

### Cost Estimate

**Free Tier** (per month):
- 10 GB storage
- 1,000,000 Class A operations (writes)
- 10,000,000 Class B operations (reads)
- Unlimited egress (downloads)

**Paid Usage** (beyond free tier):
- Storage: $0.015/GB/month
- Class A operations: $4.50/million (writes, lists)
- Class B operations: $0.36/million (reads)
- **Egress: $0** (FREE!)

**Example**: 100GB storage + 10M reads/month = **$1.50/month**
(vs AWS S3: $2.30 storage + $3.50 transfer = **$5.80/month**)

---

## Gmail SMTP Setup

Use Gmail's SMTP server to send emails for free (up to 500 emails per day).

### Why Gmail SMTP?

- **Free**: No cost for moderate usage
- **Reliable**: Google's infrastructure
- **Easy to set up**: Just need an app password
- **Good for development and small production use**

### Setup Steps

1. **Enable 2-Factor Authentication**
   - Go to [https://myaccount.google.com/security](https://myaccount.google.com/security)
   - Enable "2-Step Verification" (required for app passwords)

2. **Generate App Password**
   - Go to [https://myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
   - Select "Mail" and "Other (Custom name)"
   - Enter name: "Dear Future App"
   - Click "Generate"
   - **Copy the 16-character password** (you won't see it again)

3. **Environment Variables**
   ```bash
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-16-char-app-password
   SMTP_FROM_EMAIL=your-email@gmail.com
   SMTP_FROM_NAME=Dear Future
   SMTP_USE_TLS=true
   ```

### Sending Limits

- **500 emails per day** (rolling 24-hour period)
- **100 recipients per message**
- Sufficient for small-to-medium applications

### Alternative SMTP Providers

If you need higher volume:

1. **SendGrid**
   - Free tier: 100 emails/day
   - Paid: $15/month for 40,000 emails

2. **Mailgun**
   - Free tier: 5,000 emails/month (first 3 months)
   - Paid: $35/month for 50,000 emails

3. **AWS SES**
   - $0.10 per 1,000 emails
   - Requires AWS account

### Configuration for Other Providers

**SendGrid**:
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
```

**Mailgun**:
```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USERNAME=postmaster@your-domain.mailgun.org
SMTP_PASSWORD=your-mailgun-smtp-password
```

---

## Environment Configuration

### Development (.env.local)

```bash
# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/dear_future_dev?sslmode=disable

# No R2 or SMTP needed - uses mocks
# (or configure for testing)
```

### Staging (.env.staging)

```bash
# Database
DATABASE_URL=postgresql://user:pass@staging-db.example.com:5432/staging_db

# R2 Storage
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=staging-key-id
R2_SECRET_ACCESS_KEY=staging-secret
R2_BUCKET_NAME=dear-future-staging
R2_PUBLIC_URL=https://staging-uploads.yourdomain.com

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=staging@yourdomain.com
SMTP_PASSWORD=staging-app-password
SMTP_FROM_EMAIL=staging@yourdomain.com
SMTP_FROM_NAME=Dear Future Staging
SMTP_USE_TLS=true
```

### Production (.env.production)

```bash
# Environment
ENVIRONMENT=production

# Database (Supabase)
DATABASE_URL=postgresql://postgres:[password]@db.[project].supabase.co:5432/postgres
SUPABASE_URL=https://[project].supabase.co
SUPABASE_SERVICE_KEY=your-service-role-key

# R2 Storage
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=prod-key-id
R2_SECRET_ACCESS_KEY=prod-secret
R2_BUCKET_NAME=dear-future-prod
R2_PUBLIC_URL=https://uploads.yourdomain.com

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=prod-app-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=Dear Future
SMTP_USE_TLS=true

# Security
JWT_SECRET=your-strong-random-secret-min-32-chars
DEBUG=false
```

---

## Testing Your Setup

### Test Database Connection

```bash
# Start the application
go run main.go

# Check logs for:
# ✅ PostgreSQL database connected
```

Or test directly:
```bash
psql $DATABASE_URL -c "SELECT version();"
```

### Test R2 Storage

```go
// In your code, or create a test script
ctx := context.Background()
testFile := effects.FileUpload{
    FileName:    "test.txt",
    ContentType: "text/plain",
    Data:        []byte("Hello R2!"),
    Size:        int64(len("Hello R2!")),
}

result := r2Storage.UploadFile(ctx, testFile)
if result.IsOk() {
    fmt.Println("✅ R2 upload successful:", result.Value().URL)
} else {
    fmt.Println("❌ R2 upload failed:", result.Error())
}
```

### Test SMTP Email

```bash
# Start the application
ENVIRONMENT=production go run main.go

# Check logs for:
# ✅ SMTP email service configured
```

Or use a test script:
```go
ctx := context.Background()
deliveryInfo := message.MessageDeliveryInfo{
    MessageID:      uuid.New(),
    RecipientEmail: "your-test-email@gmail.com",
    Subject:        "Test Email",
    Content:        "This is a test from Dear Future!",
    ScheduledFor:   time.Now(),
}

result := emailService.SendMessage(ctx, deliveryInfo)
if result.IsOk() && result.Value().Status == effects.EmailStatusSent {
    fmt.Println("✅ Email sent successfully")
} else {
    fmt.Println("❌ Email failed:", result.Value().Error)
}
```

---

## Troubleshooting

### Database Connection Issues

```bash
# Error: "connection refused"
# - Check if PostgreSQL is running
# - Verify port 5432 is accessible
# - Check firewall rules

# Error: "password authentication failed"
# - Verify DATABASE_URL credentials
# - Check password doesn't contain special characters that need encoding

# Error: "SSL required"
# - For local dev, add ?sslmode=disable
# - For production, use ?sslmode=require
```

### R2 Upload Failures

```bash
# Error: "NoSuchBucket"
# - Verify bucket name is correct
# - Check bucket exists in your account

# Error: "AccessDenied"
# - Verify R2_ACCESS_KEY_ID and R2_SECRET_ACCESS_KEY
# - Check API token has "Object Read & Write" permissions
# - Ensure token is associated with the correct bucket

# Error: "InvalidAccessKeyId"
# - Regenerate API token
# - Update R2_ACCESS_KEY_ID in .env
```

### SMTP Send Failures

```bash
# Error: "username and password not accepted"
# - Verify you're using an App Password, not your Gmail password
# - Check 2FA is enabled on your Google account
# - Regenerate app password

# Error: "connection timeout"
# - Check SMTP_PORT (587 for TLS, 465 for SSL)
# - Verify firewall allows outbound SMTP
# - Try SMTP_PORT=465 with SSL

# Error: "daily sending quota exceeded"
# - Gmail limit is 500 emails/day
# - Wait 24 hours or upgrade to a paid SMTP service
```

---

## Security Best Practices

1. **Never commit secrets to git**
   - Use `.env` files (already in `.gitignore`)
   - Use environment variables in production

2. **Rotate credentials regularly**
   - R2 API tokens: every 90 days
   - Gmail app passwords: if exposed or yearly
   - Database passwords: every 90 days

3. **Use different credentials for each environment**
   - Separate tokens for dev/staging/production
   - Different email accounts for staging/production

4. **Enable monitoring**
   - Set up Cloudflare alerts for R2 usage
   - Monitor email sending quotas
   - Track database connection pools

5. **Backup your data**
   - Enable Supabase automated backups
   - Consider R2 bucket versioning
   - Store credentials in a password manager

---

## Cost Summary (Monthly)

### Minimal Production Setup
- **Supabase**: $0 (free tier: 500MB database + 1GB file storage)
- **R2 Storage**: $0 (free tier: 10GB + unlimited egress)
- **Gmail SMTP**: $0 (free: 500 emails/day)
- **Total**: **$0/month** 🎉

### Small Production (1000 users)
- **Supabase**: $25 (Pro plan: 8GB database)
- **R2 Storage**: $1.50 (100GB storage)
- **Gmail SMTP**: $0 (within 500/day limit)
- **Total**: **$26.50/month**

### Medium Production (10,000 users)
- **Supabase**: $25-100 (Pro plan)
- **R2 Storage**: $15 (1TB storage)
- **SendGrid**: $15 (40k emails/month)
- **Total**: **$55-130/month**

---

## Next Steps

1. ✅ Set up database (PostgreSQL/Supabase)
2. ✅ Configure R2 storage for file uploads
3. ✅ Set up Gmail SMTP for emails
4. 🚀 Deploy your application!

For deployment guides, see:
- [Deployment Guide](deployment-guide.md)
- [Configuration Guide](configuration-guide.md)
