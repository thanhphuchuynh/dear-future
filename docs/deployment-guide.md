# Deployment Guide - Migration-Ready Architecture

## Overview

This guide covers deployment strategies for the Dear Future application, designed to start ultra-cheap with Lambda and migrate seamlessly to ECS or AKS as needed.

## Deployment Options Comparison

| Platform | Cost/Month | Complexity | Use Case |
|----------|------------|------------|----------|
| **AWS Lambda** | $0-5 | Low | Start here - MVP, low traffic |
| **AWS ECS Fargate** | $30-50 | Medium | Scale up - consistent traffic |
| **Azure AKS** | $100-300 | High | Enterprise - high traffic, multi-cloud |

## Phase 1: Lambda Deployment (Start Here)

### Prerequisites
- AWS Account with CLI configured
- Go 1.21+ installed
- SAM CLI installed
- Docker installed

### Project Structure for Lambda
```
dear-future/
├── cmd/
│   └── lambda/
│       └── main.go          # Lambda entry point
├── template.yaml            # SAM template
├── Makefile                # Build automation
└── scripts/
    └── deploy.sh           # Deployment script
```

### SAM Template (`template.yaml`)
```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31

Globals:
  Function:
    Timeout: 30
    MemorySize: 512
    Runtime: provided.al2
    Architectures:
      - x86_64

Parameters:
  Environment:
    Type: String
    Default: dev
    AllowedValues: [dev, staging, prod]

Resources:
  DearFutureAPI:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: ./
      Handler: bootstrap
      Environment:
        Variables:
          ENVIRONMENT: !Ref Environment
          DATABASE_URL: !Ref DatabaseURL
          JWT_SECRET: !Ref JWTSecret
          S3_BUCKET: !Ref AttachmentsBucket
      Events:
        ApiEvent:
          Type: Api
          Properties:
            Path: /{proxy+}
            Method: ANY
            RestApiId: !Ref DearFutureApiGateway

  DearFutureApiGateway:
    Type: AWS::Serverless::Api
    Properties:
      StageName: !Ref Environment
      Cors:
        AllowMethods: "'*'"
        AllowHeaders: "'*'"
        AllowOrigin: "'*'"

  MessageScheduler:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: ./
      Handler: scheduler
      Events:
        ScheduleEvent:
          Type: Schedule
          Properties:
            Schedule: rate(1 hour)

  AttachmentsBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub "dear-future-attachments-${Environment}"
      CorsConfiguration:
        CorsRules:
          - AllowedHeaders: ["*"]
            AllowedMethods: [GET, POST, PUT]
            AllowedOrigins: ["*"]

Outputs:
  ApiUrl:
    Description: API Gateway endpoint URL
    Value: !Sub "https://${DearFutureApiGateway}.execute-api.${AWS::Region}.amazonaws.com/${Environment}"
```

### Lambda Entry Point
```go
// cmd/lambda/main.go
package main

import (
    "context"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/awslabs/aws-lambda-go-api-proxy/gin"
    "your-app/pkg/composition"
)

var ginLambda *ginadapter.GinLambda

func init() {
    app, err := composition.NewApp()
    if err != nil {
        panic(err)
    }
    
    router := app.SetupRoutes()
    ginLambda = ginadapter.New(router)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
    lambda.Start(Handler)
}
```

### Deployment Commands
```bash
# Build and deploy
make deploy-lambda

# Or manually:
sam build
sam deploy --guided  # First time
sam deploy           # Subsequent deployments
```

### Makefile for Lambda
```makefile
.PHONY: build-lambda deploy-lambda test

build-lambda:
	GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/lambda/main.go
	zip lambda-deployment.zip bootstrap

deploy-lambda: build-lambda
	sam deploy --parameter-overrides Environment=dev

test:
	go test ./...

local-api:
	sam local start-api --port 8080
```

## Phase 2: ECS Deployment (Scale Up)

### When to Migrate to ECS
- Lambda costs exceed $20/month
- Cold start latency becomes problematic
- Need background processing longer than 15 minutes
- Consistent traffic patterns

### ECS Project Structure
```
dear-future/
├── cmd/
│   └── server/
│       └── main.go          # HTTP server entry point
├── deployments/
│   └── ecs/
│       ├── Dockerfile
│       ├── docker-compose.yml
│       ├── task-definition.json
│       └── service.yaml
└── scripts/
    └── deploy-ecs.sh
```

### Multi-Stage Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./server"]
```

### ECS Task Definition
```json
{
  "family": "dear-future-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "dear-future-api",
      "image": "your-account.dkr.ecr.region.amazonaws.com/dear-future:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENVIRONMENT",
          "value": "production"
        },
        {
          "name": "PORT",
          "value": "8080"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:ssm:region:account:parameter/dear-future/database-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/dear-future",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### ECS Deployment Script
```bash
#!/bin/bash
# scripts/deploy-ecs.sh

set -e

# Build and push image
docker build -t dear-future .
docker tag dear-future:latest $ECR_REGISTRY/dear-future:latest
docker push $ECR_REGISTRY/dear-future:latest

# Update ECS service
aws ecs update-service \
    --cluster dear-future-cluster \
    --service dear-future-api \
    --force-new-deployment

echo "ECS deployment completed"
```

## Phase 3: AKS Deployment (Enterprise Scale)

### When to Migrate to AKS
- Need multi-cloud strategy
- Require advanced orchestration features
- Team has Kubernetes expertise
- Traffic exceeds 100K+ users

### Kubernetes Manifests

#### Deployment
```yaml
# deployments/k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dear-future-api
  labels:
    app: dear-future-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dear-future-api
  template:
    metadata:
      labels:
        app: dear-future-api
    spec:
      containers:
      - name: api
        image: dearfuture.azurecr.io/dear-future:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: PORT
          value: "8080"
        envFrom:
        - secretRef:
            name: dear-future-secrets
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### Service
```yaml
# deployments/k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: dear-future-api-service
spec:
  selector:
    app: dear-future-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

#### Ingress
```yaml
# deployments/k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dear-future-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.dearfuture.com
    secretName: dear-future-tls
  rules:
  - host: api.dearfuture.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: dear-future-api-service
            port:
              number: 80
```

### AKS Deployment Commands
```bash
# Build and push to Azure Container Registry
az acr build --registry dearfuture --image dear-future:latest .

# Deploy to AKS
kubectl apply -f deployments/k8s/

# Scale deployment
kubectl scale deployment dear-future-api --replicas=5

# Rolling update
kubectl set image deployment/dear-future-api api=dearfuture.azurecr.io/dear-future:v2
```

## CI/CD Pipeline

### GitHub Actions for Multi-Platform Deployment
```yaml
# .github/workflows/deploy.yml
name: Deploy Dear Future

on:
  push:
    branches: [main, staging]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - run: go test ./...

  deploy-lambda:
    if: github.ref == 'refs/heads/staging'
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: aws-actions/setup-sam@v2
    - run: sam build && sam deploy --no-confirm-changeset --no-fail-on-empty-changeset

  deploy-ecs:
    if: github.ref == 'refs/heads/main'
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1
    - run: ./scripts/deploy-ecs.sh
```

## Migration Checklist

### Lambda → ECS Migration
- [ ] Create ECS cluster and task definition
- [ ] Set up Application Load Balancer
- [ ] Configure auto-scaling policies
- [ ] Update DNS to point to ALB
- [ ] Monitor performance and costs
- [ ] Decommission Lambda functions

### ECS → AKS Migration
- [ ] Set up AKS cluster
- [ ] Create Kubernetes manifests
- [ ] Set up ingress controller
- [ ] Configure monitoring (Prometheus/Grafana)
- [ ] Migrate traffic gradually
- [ ] Decommission ECS resources

## Monitoring & Observability

### Lambda Monitoring
- CloudWatch Logs and Metrics
- X-Ray tracing (optional)
- Custom metrics via CloudWatch

### ECS Monitoring
- CloudWatch Container Insights
- Application Load Balancer metrics
- Custom application metrics

### AKS Monitoring
- Azure Monitor for containers
- Prometheus + Grafana
- Jaeger for distributed tracing

## Cost Optimization

### Lambda Costs
- Use ARM architecture for 20% cost savings
- Optimize memory allocation
- Implement connection pooling

### ECS Costs
- Use Spot instances for non-critical workloads
- Right-size container resources
- Use reserved capacity for predictable workloads

### AKS Costs
- Use Azure Spot VMs
- Implement horizontal pod autoscaling
- Use Azure Reserved VM Instances

## Security Considerations

### All Platforms
- Use least privilege IAM/RBAC policies
- Encrypt data in transit and at rest
- Regular security updates
- Implement WAF rules

### Platform-Specific
- **Lambda**: VPC configuration for database access
- **ECS**: Security groups and task roles
- **AKS**: Network policies and pod security standards

This deployment guide ensures smooth migration between platforms while maintaining security, performance, and cost-effectiveness at each stage.