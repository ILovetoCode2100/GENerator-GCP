# Virtuoso API CLI Helm Chart - Quick Start Guide

## Prerequisites

1. Kubernetes cluster (1.19+)
2. Helm 3 installed
3. kubectl configured
4. Virtuoso API token
5. Virtuoso Organization ID

## Quick Installation

### 1. Clone the repository

```bash
git clone https://github.com/virtuoso/virtuoso-api-cli.git
cd virtuoso-api-cli/deployment/helm
```

### 2. Install for Development

```bash
./install.sh \
  --environment dev \
  --api-token "your-dev-api-token" \
  --org-id "2242-dev" \
  install
```

### 3. Install for Production

```bash
./install.sh \
  --namespace virtuoso-prod \
  --release virtuoso-api-prod \
  --environment production \
  --api-token "your-prod-api-token" \
  --org-id "2242" \
  install
```

## Verify Installation

```bash
# Check pods
kubectl get pods -n virtuoso-api

# Check services
kubectl get svc -n virtuoso-api

# Check ingress (if enabled)
kubectl get ingress -n virtuoso-api

# View logs
kubectl logs -l app.kubernetes.io/name=virtuoso-api-cli -n virtuoso-api
```

## Access the Application

### Port Forward (for testing)

```bash
kubectl port-forward -n virtuoso-api svc/virtuoso-api-cli 8080:80
# Access at http://localhost:8080
```

### Via Ingress (if enabled)

Access the URL shown in the ingress configuration.

## Common Operations

### Update API Token

```bash
kubectl create secret generic virtuoso-api-cli-secret \
  --from-literal=API_TOKEN="new-token" \
  --namespace virtuoso-api \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods to pick up new secret
kubectl rollout restart deployment/virtuoso-api-cli -n virtuoso-api
```

### Scale the Deployment

```bash
# Manual scaling
kubectl scale deployment/virtuoso-api-cli --replicas=5 -n virtuoso-api

# Or enable autoscaling
./install.sh \
  --namespace virtuoso-api \
  --release virtuoso-api-cli \
  upgrade
```

### View Metrics (if Prometheus is enabled)

```bash
kubectl port-forward -n virtuoso-api svc/virtuoso-api-cli 8080:80
curl http://localhost:8080/metrics
```

## Troubleshooting

### Pod not starting

```bash
# Check pod events
kubectl describe pod -l app.kubernetes.io/name=virtuoso-api-cli -n virtuoso-api

# Check logs
kubectl logs -l app.kubernetes.io/name=virtuoso-api-cli -n virtuoso-api --previous
```

### Configuration issues

```bash
# View current configuration
kubectl get configmap virtuoso-api-cli-config -n virtuoso-api -o yaml

# View secrets (without values)
kubectl get secret virtuoso-api-cli-secret -n virtuoso-api -o yaml
```

### Redis connection issues

```bash
# Check Redis pod
kubectl get pod -l app.kubernetes.io/component=redis -n virtuoso-api

# Test Redis connection
kubectl exec -it deployment/virtuoso-api-cli -n virtuoso-api -- nc -zv virtuoso-api-cli-redis 6379
```

## Uninstall

```bash
./install.sh \
  --namespace virtuoso-api \
  --release virtuoso-api-cli \
  uninstall
```

## Next Steps

1. Configure monitoring with ServiceMonitor
2. Set up alerts with PrometheusRule
3. Enable external secrets for better security
4. Configure network policies
5. Set up backup procedures

For more details, see the full [README.md](virtuoso-api-cli/README.md).
