# Deployment Guide

> Comprehensive deployment guide for OctAi

## Table of Contents

- [Docker Deployment](#docker-deployment)
- [Docker Compose](#docker-compose)
- [Kubernetes](#kubernetes)
- [Environment Variables](#environment-variables)
- [Production Considerations](#production-considerations)

---

## Docker Deployment

### Quick Start

```bash
# Pull the latest image
docker pull docker.io/sipeed/octai:latest

# Run the gateway mode
docker run -d --name octai \
  -v ~/.octai:/root/.octai \
  -p 18790:18790 \
  docker.io/sipeed/octai:latest gateway
```

### Available Images

| Image | Description | Size |
|-------|-------------|------|
| `octai:latest` | Minimal image for gateway | ~15MB |
| `octai:launcher` | Web console + gateway | ~25MB |
| `octai:full` | Full MCP support | ~30MB |

### Building Custom Images

```bash
# Standard build
docker build -t octai:latest -f docker/Dockerfile .

# Full MCP support
docker build -t octai:full -f docker/Dockerfile.full .
```

---

## Docker Compose

### Basic Configuration

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  octai-gateway:
    image: docker.io/sipeed/octai:latest
    container_name: octai-gateway
    restart: unless-stopped
    environment:
      - AIBHQ_GATEWAY_HOST=0.0.0.0
    volumes:
      - ./data:/root/.octai
    ports:
      - "18790:18790"
```

### Full Stack with Multiple Services

```yaml
version: '3.8'

services:
  octai-gateway:
    image: docker.io/sipeed/octai:latest
    container_name: octai-gateway
    restart: unless-stopped
    environment:
      - AIBHQ_GATEWAY_HOST=0.0.0.0
    volumes:
      - ./config:/root/.octai/config.json:ro
      - octai-workspace:/root/.octai/workspace
    ports:
      - "18790:18790"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:18790/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Optional: Redis for session caching
  redis:
    image: redis:7-alpine
    container_name: octai-redis
    restart: unless-stopped
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"

  # Optional: Reverse proxy
  nginx:
    image: nginx:alpine
    container_name: octai-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - octai-gateway

volumes:
  octai-workspace:
  redis-data:
```

### Start Services

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f octai-gateway

# Stop services
docker-compose down
```

---

## Kubernetes

### Basic Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: octai
  labels:
    app: octai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: octai
  template:
    metadata:
      labels:
        app: octai
    spec:
      containers:
        - name: octai
          image: docker.io/sipeed/octai:latest
          args: ["gateway"]
          env:
            - name: AIBHQ_GATEWAY_HOST
              value: "0.0.0.0"
            - name: AIBHQ_CONFIG
              value: "/config/config.json"
          ports:
            - containerPort: 18790
          volumeMounts:
            - name: config
              mountPath: /config
              readOnly: true
            - name: workspace
              mountPath: /root/.octai/workspace
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "500m"
      volumes:
        - name: config
          configMap:
            name: octai-config
        - name: workspace
          persistentVolumeClaim:
            claimName: octai-workspace
            readOnly: false
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: octai-config
data:
  config.json: |
    {
      "model_list": [
        {
          "model_name": "gpt-4o-mini",
          "model": "openai/gpt-4o-mini",
          "api_key": "YOUR_API_KEY"
        }
      ],
      "agents": {
        "defaults": {
          "model_name": "gpt-4o-mini"
        }
      }
    }
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: octai-workspace
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: octai
spec:
  selector:
    app: octai
  ports:
    - port: 18790
      targetPort: 18790
```

### Deploy to Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s-octai.yaml

# Check status
kubectl get pods -l app=octai

# View logs
kubectl logs -f deployment/octai
```

### Horizontal Scaling

```yaml
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
```

---

## Environment Variables

### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AIBHQ_CONFIG` | `~/.octai/config.json` | Path to configuration file |
| `AIBHQ_HOME` | `~/.octai` | Root directory for data |
| `AIBHQ_GATEWAY_HOST` | `127.0.0.1` | Gateway bind address |
| `AIBHQ_GATEWAY_PORT` | `18790` | Gateway HTTP port |

### Agent Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AIBHQ_AGENTS_DEFAULTS_MODEL_NAME` | - | Default model name |
| `AIBHQ_AGENTS_DEFAULTS_WORKSPACE` | `~/.octai/workspace` | Workspace directory |
| `AIBHQ_AGENTS_DEFAULTS_MAX_TOKENS` | `4096` | Maximum response tokens |
| `AIBHQ_AGENTS_DEFAULTS_TEMPERATURE` | `0.7` | Response temperature |
| `AIBHQ_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE` | `true` | Sandbox mode |

### Feature Flags

| Variable | Default | Description |
|----------|---------|-------------|
| `AIBHQ_HEARTBEAT_ENABLED` | `true` | Enable periodic tasks |
| `AIBHQ_HEARTBEAT_INTERVAL` | `30` | Heartbeat interval (minutes) |
| `AIBHQ_TOOLS_EXEC_ALLOW_REMOTE` | `false` | Allow exec from remote channels |
| `AIBHQ_TOOLS_WEB_ENABLED` | `true` | Enable web search tools |

### Docker-Specific

| Variable | Default | Description |
|----------|---------|-------------|
| `AIBHQ_DOCKER_DATA_DIR` | `/root/.octai` | Data directory in container |

### Example Docker Compose Environment

```yaml
services:
  octai:
    environment:
      - AIBHQ_GATEWAY_HOST=0.0.0.0
      - AIBHQ_AGENTS_DEFAULTS_MODEL_NAME=gpt-4o-mini
      - AIBHQ_HEARTBEAT_ENABLED=true
      - AIBHQ_HEARTBEAT_INTERVAL=15
```

---

## Production Considerations

### Security

#### 1. Network Security

```yaml
# Restrict API access to internal network only
services:
  octai:
    networks:
      - internal
    expose:
      - "18790"

networks:
  internal:
    internal: true
```

#### 2. Secrets Management

**Using Docker Secrets:**

```bash
# Create secret
echo "sk-your-api-key" | docker secret create octai-api-key -

# Use in compose
services:
  octai:
    secrets:
      - octai-api-key
    environment:
      - OPENAI_API_KEY_FILE=/run/secrets/octai-api-key
```

**Using Kubernetes Secrets:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: octai-secrets
type: Opaque
stringData:
  config.json: |
    {
      "model_list": [
        {
          "model_name": "gpt-4o-mini",
          "model": "openai/gpt-4o-mini",
          "api_key": "sk-your-api-key"
        }
      ]
    }
```

#### 3. HTTPS/TLS

```yaml
# nginx.conf for TLS termination
server {
    listen 443 ssl;
    server_name octai.example.com;

    ssl_certificate /etc/nginx/ssl/tls.crt;
    ssl_certificate_key /etc/nginx/ssl/tls.key;

    location / {
        proxy_pass http://octai-gateway:18790;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Performance

#### Resource Limits

```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

#### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: octai-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: octai
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

### High Availability

#### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 18790
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /health
    port: 18790
  initialDelaySeconds: 5
  periodSeconds: 10
```

#### Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: octai-pdb
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: octai
```

### Monitoring

#### Prometheus Metrics

```yaml
# Enable metrics endpoint
environment:
  - AIBHQ_METRICS_ENABLED=true
  - AIBHQ_METRICS_PORT=9090
```

#### Log Aggregation

```yaml
# Fluent Bit configuration
[INPUT]
    Name              tail
    Path              /var/log/containers/*.log
    Parser            docker
    Tag               octai

[OUTPUT]
    Match              *
    Name               stdout
```

### Backup Strategy

#### Volume Snapshots

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: octai-snapshot-class
driver: csi-driver
parameters:
  type: snap
deletionPolicy: Delete
```

#### Automated Backups

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: octai-backup
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: backup
              image: backup-tool:latest
              command: ["backup.sh", "/data", "/backup"]
          restartPolicy: OnFailure
```
