# Admin API Deployment

This directory contains deployment configurations and scripts for the Admin API service.

## Overview

The Admin API provides a secure HTTP interface for managing work queue operations including:
- Queue statistics and monitoring
- Job inspection and management
- Dead letter queue operations
- System benchmarking

## Structure

```
deployments/
├── docker/
│   ├── Dockerfile.admin-api    # Production Docker image
│   └── docker-compose.yaml     # Local development setup
├── kubernetes/
│   ├── admin-api-deployment.yaml  # K8s deployment manifests
│   └── monitoring.yaml            # Prometheus monitoring setup
└── scripts/
    └── deploy-staging.sh          # Staging deployment script
```

## Local Development

### Using Docker Compose

1. Start the services:
   ```bash
   cd deployments/docker
   docker-compose up -d
   ```

2. Test the API:
   ```bash
   # Health check
   curl http://localhost:8080/health

   # Get stats (requires token)
   curl -H "X-API-Token: local-dev-token" http://localhost:8080/api/v1/stats
   ```

3. Access monitoring:
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

## Staging Deployment

### Prerequisites

- kubectl configured for staging cluster
- Docker registry access
- Required environment variables set

### Deploy to Staging

```bash
# Set environment variables
export DOCKER_REGISTRY=your-registry.com
export IMAGE_TAG=v1.0.0

# Run deployment
./deployments/scripts/deploy-staging.sh
```

### Rollback

```bash
# Rollback to previous version
./deployments/scripts/deploy-staging.sh --rollback
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis server hostname | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_PASSWORD` | Redis password | `""` |
| `API_TOKEN_1` | Service account token | Required |
| `API_TOKEN_2` | Readonly account token | Required |
| `LOG_LEVEL` | Logging level | `info` |

### Secrets Management

Secrets are managed via Kubernetes secrets:

```bash
# Create API tokens secret
kubectl create secret generic admin-api-secrets \
  --from-literal=api-token-1=your-service-token \
  --from-literal=api-token-2=your-readonly-token \
  --from-literal=redis-password=your-redis-password \
  -n work-queue
```

## Monitoring

The deployment includes comprehensive monitoring setup:

### Metrics

- HTTP request metrics (rate, latency, errors)
- Authentication and authorization metrics
- Rate limiting metrics
- Audit log metrics
- System resource metrics

### Alerts

- High error rate (>5%)
- Service downtime
- High latency (>1s p95)
- Memory usage (>90%)
- Certificate expiry
- Audit log failures

### Dashboards

Grafana dashboards are automatically provisioned showing:
- Request throughput and latency
- Error rates by endpoint
- Resource utilization
- Rate limiting statistics

## Security

### Authentication

All API endpoints except `/health` require authentication via `X-API-Token` header.

### Authorization

Two token types are supported:
- Service tokens: Full read/write access
- Readonly tokens: Read-only access

### Rate Limiting

- 100 requests per minute per client
- Burst allowance of 20 requests
- 429 responses for exceeded limits

### Audit Logging

All destructive operations are logged to audit files including:
- Timestamp and user identification
- Action performed
- Request parameters
- Response status

### Network Security

- TLS encryption for all external traffic
- Internal service mesh encryption
- Network policies restricting pod-to-pod communication

## Health Checks

The service provides multiple health check endpoints:

- `/health`: Basic liveness check
- Kubernetes probes: Liveness and readiness probes
- Prometheus metrics: Detailed health metrics

## Scaling

### Horizontal Pod Autoscaling

The deployment includes HPA configuration:
- Min replicas: 2
- Max replicas: 10
- CPU target: 70%
- Memory target: 80%

### Pod Disruption Budget

PDB ensures at least 1 pod remains available during:
- Rolling updates
- Node maintenance
- Voluntary disruptions

## Troubleshooting

### Common Issues

1. **Service not starting**
   - Check Redis connectivity
   - Verify API tokens are set
   - Review logs: `kubectl logs -n work-queue deployment/admin-api`

2. **Authentication failures**
   - Verify tokens in secrets
   - Check token format and permissions
   - Review audit logs for failed attempts

3. **High latency**
   - Check Redis performance
   - Review rate limiting settings
   - Scale up replicas if needed

### Debug Commands

```bash
# Check pod status
kubectl get pods -n work-queue -l app=admin-api

# View logs
kubectl logs -n work-queue -l app=admin-api --tail=100

# Port forward for local testing
kubectl port-forward -n work-queue service/admin-api 8080:80

# Check metrics
kubectl port-forward -n work-queue service/admin-api 8080:80
curl http://localhost:8080/metrics
```

## Production Deployment

**⚠️ IMPORTANT**: This deployment is configured for staging only. For production deployment:

1. Update ingress hostname
2. Configure production Redis endpoints
3. Set production API tokens
4. Update resource limits
5. Configure backup procedures
6. Set up log aggregation
7. Configure additional security scanning

Do NOT deploy to production without proper security review and testing.