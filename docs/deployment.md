# Deployment Guide

This guide covers various deployment options for OmniTranscripts, from local development to production-scale deployments.

## Quick Deployment Options

### 1. Docker (Recommended)

The fastest way to get OmniTranscripts running in any environment.

#### Using Pre-built Image (Coming Soon)
```bash
docker run -d \
  --name videotranscript \
  -p 3000:3000 \
  -e API_KEY=your-secure-api-key \
  -e WORK_DIR=/tmp/videotranscript \
  wilmoore/videotranscript:latest
```

#### Build from Source
```bash
# Clone the repository
git clone https://github.com/wilmoore/OmniTranscripts.git
cd OmniTranscripts

# Build the Docker image
docker build -t videotranscript-app .

# Run the container
docker run -d \
  --name videotranscript \
  -p 3000:3000 \
  --env-file .env \
  -v ./transcripts:/app/transcripts \
  videotranscript-app
```

### 2. Encore.dev (Production Ready)

Deploy to production with a single command using Encore.dev's cloud platform.

```bash
# Install Encore CLI
curl -L https://encore.dev/install.sh | bash

# Deploy to production
encore deploy --env production

# Or deploy to staging
encore deploy --env staging
```

**Benefits:**
- Automatic scaling and load balancing
- Built-in monitoring and observability
- Managed PostgreSQL database
- Zero-downtime deployments
- Built-in authentication and authorization

## Production Deployment

### Prerequisites

#### System Requirements
- **CPU**: 2+ cores (4+ recommended for high throughput)
- **Memory**: 4GB+ RAM (8GB+ recommended)
- **Storage**: 20GB+ SSD for temporary file processing
- **Network**: Reliable internet connection for YouTube downloads

#### Required Dependencies
- **Go 1.23+**: Application runtime
- **yt-dlp**: YouTube video downloading
- **FFmpeg**: Audio processing
- **whisper.cpp**: AI transcription

### Environment Configuration

#### Production Environment Variables
```env
# Server Configuration
PORT=3000
API_KEY=your-secure-production-api-key-here

# Processing Configuration
WORK_DIR=/var/lib/videotranscript
MAX_VIDEO_LENGTH=1800
FREE_JOB_LIMIT=5

# Database (for Encore.dev deployments)
DATABASE_URL=postgresql://user:pass@host:5432/videotranscript

# Redis (for job queue scaling)
REDIS_URL=redis://user:pass@host:6379/0

# Monitoring (optional)
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project
DATADOG_API_KEY=your-datadog-api-key

# Webhook Configuration (optional)
WEBHOOK_URL=https://your-app.com/webhooks/transcription
WEBHOOK_SECRET=your-webhook-secret
WEBHOOK_EVENTS=job.started,job.completed,job.failed
```

### Cloud Platform Deployments

#### AWS ECS Deployment

**Task Definition:**
```json
{
  "family": "videotranscript-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::account:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::account:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "videotranscript",
      "image": "your-account.dkr.ecr.region.amazonaws.com/videotranscript:latest",
      "portMappings": [
        {
          "containerPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {"name": "PORT", "value": "3000"},
        {"name": "API_KEY", "value": "your-api-key"},
        {"name": "WORK_DIR", "value": "/tmp/videotranscript"}
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/videotranscript",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

**Service Definition:**
```json
{
  "serviceName": "videotranscript-service",
  "cluster": "production",
  "taskDefinition": "videotranscript-app",
  "desiredCount": 3,
  "launchType": "FARGATE",
  "networkConfiguration": {
    "awsvpcConfiguration": {
      "subnets": ["subnet-12345", "subnet-67890"],
      "securityGroups": ["sg-abcdef"],
      "assignPublicIp": "ENABLED"
    }
  },
  "loadBalancers": [
    {
      "targetGroupArn": "arn:aws:elasticloadbalancing:region:account:targetgroup/videotranscript-tg",
      "containerName": "videotranscript",
      "containerPort": 3000
    }
  ]
}
```

#### Google Cloud Run Deployment

```bash
# Build and push image
gcloud builds submit --tag gcr.io/PROJECT_ID/videotranscript

# Deploy to Cloud Run
gcloud run deploy videotranscript \
  --image gcr.io/PROJECT_ID/videotranscript \
  --platform managed \
  --region us-central1 \
  --set-env-vars API_KEY=your-api-key \
  --set-env-vars WORK_DIR=/tmp/videotranscript \
  --memory 2Gi \
  --cpu 2 \
  --max-instances 10 \
  --allow-unauthenticated
```

#### Azure Container Instances

```bash
az container create \
  --resource-group videotranscript-rg \
  --name videotranscript-app \
  --image your-registry/videotranscript:latest \
  --cpu 2 \
  --memory 4 \
  --ports 3000 \
  --environment-variables \
    API_KEY=your-api-key \
    WORK_DIR=/tmp/videotranscript \
  --dns-name-label videotranscript-unique
```

### Kubernetes Deployment

#### Deployment Manifest
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: videotranscript-app
  labels:
    app: videotranscript
spec:
  replicas: 3
  selector:
    matchLabels:
      app: videotranscript
  template:
    metadata:
      labels:
        app: videotranscript
    spec:
      containers:
      - name: videotranscript
        image: videotranscript:latest
        ports:
        - containerPort: 3000
        env:
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: videotranscript-secrets
              key: api-key
        - name: WORK_DIR
          value: "/tmp/videotranscript"
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        volumeMounts:
        - name: temp-storage
          mountPath: /tmp/videotranscript
      volumes:
      - name: temp-storage
        emptyDir:
          sizeLimit: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: videotranscript-service
spec:
  selector:
    app: videotranscript
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
---
apiVersion: v1
kind: Secret
metadata:
  name: videotranscript-secrets
type: Opaque
data:
  api-key: eW91ci1iYXNlNjQtZW5jb2RlZC1hcGkta2V5
```

#### Horizontal Pod Autoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: videotranscript-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: videotranscript-app
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Traditional Server Deployment

#### Ubuntu/Debian Setup
```bash
# Install dependencies
sudo apt update
sudo apt install -y golang-go ffmpeg python3-pip

# Install yt-dlp
pip3 install yt-dlp

# Install whisper.cpp
git clone https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp
make
sudo cp main /usr/local/bin/whisper.cpp
bash ./models/download-ggml-model.sh base.en
sudo mkdir -p /usr/local/share/whisper
sudo cp models/ggml-base.en.bin /usr/local/share/whisper/

# Deploy application
git clone https://github.com/wilmoore/OmniTranscripts.git
cd OmniTranscripts
go build -o videotranscript-app

# Create systemd service
sudo tee /etc/systemd/system/videotranscript.service > /dev/null <<EOF
[Unit]
Description=OmniTranscripts API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/videotranscript
ExecStart=/opt/videotranscript/videotranscript-app
Restart=always
RestartSec=10
Environment=PORT=3000
Environment=API_KEY=your-api-key
Environment=WORK_DIR=/tmp/videotranscript

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable videotranscript
sudo systemctl start videotranscript
```

#### Nginx Reverse Proxy
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Increase timeouts for long-running transcription jobs
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 300s;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://127.0.0.1:3000/health;
        access_log off;
    }
}

# SSL configuration (use certbot for Let's Encrypt)
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Database Setup

### PostgreSQL (for Encore.dev or manual setup)

#### Installation
```bash
# Ubuntu/Debian
sudo apt install postgresql postgresql-contrib

# macOS
brew install postgresql

# Start service
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### Database Setup
```sql
-- Create database and user
CREATE DATABASE videotranscript;
CREATE USER videotranscript_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE videotranscript TO videotranscript_user;

-- Connect to database
\c videotranscript

-- Run migrations (manual setup)
-- Copy and execute SQL from transcribe/migrations/*.up.sql files
```

### Redis (for job queue scaling)

#### Installation
```bash
# Ubuntu/Debian
sudo apt install redis-server

# macOS
brew install redis

# Start service
sudo systemctl start redis
sudo systemctl enable redis
```

#### Configuration
```redis
# /etc/redis/redis.conf
bind 127.0.0.1
port 6379
requirepass your-secure-redis-password
maxmemory 256mb
maxmemory-policy allkeys-lru
```

## Monitoring & Observability

### Health Checks

#### Kubernetes Liveness/Readiness Probes
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3000
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 5
```

#### Load Balancer Health Checks
```bash
# AWS ALB Health Check
curl -f http://your-app.com/health || exit 1

# Google Cloud Load Balancer
curl -f http://your-app.com/health
```

### Application Metrics

#### Prometheus Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'videotranscript'
    static_configs:
      - targets: ['localhost:3000']
    metrics_path: /metrics
    scrape_interval: 5s
```

#### Grafana Dashboard
Import dashboard from `docs/monitoring/grafana-dashboard.json` (to be created) for pre-configured metrics visualization.

### Logging

#### Centralized Logging (ELK Stack)
```yaml
# docker-compose.logging.yml
version: '3.8'
services:
  videotranscript:
    image: videotranscript:latest
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "5"
        labels: "service=videotranscript"

  filebeat:
    image: docker.elastic.co/beats/filebeat:7.15.0
    volumes:
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

## Performance Tuning

### Application Optimization

#### Worker Pool Configuration
```go
// Adjust based on server resources
const (
    MaxWorkers = 4  // Number of concurrent transcription jobs
    QueueSize  = 100 // Maximum queued jobs
)
```

#### Memory Management
```bash
# Set Go garbage collection target
export GOGC=100

# Limit memory usage
ulimit -v 4194304  # 4GB virtual memory limit
```

### Infrastructure Optimization

#### CPU Optimization
- Use CPU-optimized instances for transcription workloads
- Consider ARM-based instances for cost savings
- Enable hyperthreading if available

#### Storage Optimization
```bash
# Use SSD for temporary file processing
mkdir -p /mnt/ssd/videotranscript
export WORK_DIR=/mnt/ssd/videotranscript

# Regularly clean temporary files
find $WORK_DIR -type f -mtime +1 -delete
```

#### Network Optimization
- Use CDN for static assets
- Enable compression in reverse proxy
- Optimize DNS resolution

## Security Considerations

### API Security
```nginx
# Rate limiting in Nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

server {
    location /transcribe {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://backend;
    }
}
```

### Container Security
```dockerfile
# Use non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser
```

### Network Security
- Use HTTPS/TLS for all communications
- Implement proper firewall rules
- Use VPC/private networks for internal communication
- Regular security updates for dependencies

## Scaling Strategies

### Horizontal Scaling
1. **Load Balancer**: Distribute requests across multiple instances
2. **Database Sharding**: Partition data across multiple databases
3. **Cache Layer**: Redis cluster for session and job data
4. **CDN**: Distribute static assets globally

### Vertical Scaling
1. **CPU**: Increase cores for parallel processing
2. **Memory**: More RAM for larger audio files
3. **Storage**: Faster SSD for temporary file operations
4. **Network**: Higher bandwidth for video downloads

### Auto-scaling Rules
```yaml
# Example auto-scaling configuration
rules:
  - metric: cpu_utilization
    threshold: 70%
    action: scale_up
    cooldown: 300s

  - metric: queue_length
    threshold: 50
    action: scale_up
    cooldown: 180s

  - metric: cpu_utilization
    threshold: 30%
    action: scale_down
    cooldown: 600s
```

This deployment guide provides comprehensive options for running OmniTranscripts in any environment, from development to enterprise-scale production deployments.