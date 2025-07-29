# Virtuoso API CLI - Kubernetes Deployment

This directory contains the base Kubernetes manifests for deploying the Virtuoso API CLI in a production environment.

## Prerequisites

1. Kubernetes cluster v1.21+
2. NGINX Ingress Controller installed
3. cert-manager installed (for TLS certificates)
4. kubectl and kustomize CLI tools

## Quick Start

1. **Update Configuration**

   - Edit `secret.yaml` with your actual credentials
   - Update `ingress.yaml` with your domain name
   - Adjust resource limits in `deployment.yaml` if needed

2. **Deploy using kubectl**

   ```bash
   kubectl apply -k .
   ```

3. **Verify Deployment**
   ```bash
   kubectl -n virtuoso-api get all
   kubectl -n virtuoso-api get ingress
   ```

## Components

### Core Services

- **API Deployment**: 3 replicas with health checks and resource limits
- **Redis StatefulSet**: Persistent cache with password authentication
- **Service**: ClusterIP service with session affinity

### Configuration

- **ConfigMap**: Non-sensitive configuration values
- **Secret**: Sensitive credentials (must be updated before deployment)

### Security

- **NetworkPolicy**: Strict network isolation rules
- **ServiceAccount**: RBAC permissions for API pods
- **PodDisruptionBudget**: Ensures high availability during updates

### Ingress

- **TLS Termination**: Automatic certificate management
- **Rate Limiting**: Protection against abuse
- **CORS Headers**: Configured for API access
- **Security Headers**: XSS, clickjacking protection

## Customization

### Environment-Specific Overlays

Create overlays for different environments:

```bash
deployment/kubernetes/
├── base/
│   └── (these files)
└── overlays/
    ├── development/
    ├── staging/
    └── production/
```

### Scaling

Adjust replicas in `deployment.yaml`:

```yaml
spec:
  replicas: 5 # Increase for higher load
```

### Resource Limits

Modify resource requests/limits based on your needs:

```yaml
resources:
  requests:
    cpu: 1000m
    memory: 2Gi
  limits:
    cpu: 2000m
    memory: 4Gi
```

## Monitoring

The deployment includes:

- Prometheus metrics endpoint at `/metrics`
- Health check at `/health`
- Readiness check at `/ready`

## Security Considerations

1. **Secrets Management**: Consider using:

   - Sealed Secrets
   - External Secrets Operator
   - HashiCorp Vault

2. **Network Policies**: Adjust based on your cluster setup
3. **Pod Security Standards**: Runs as non-root user
4. **TLS**: Enforced for all ingress traffic

## Troubleshooting

### Check Pod Status

```bash
kubectl -n virtuoso-api describe pods
kubectl -n virtuoso-api logs -l app=virtuoso-api-cli
```

### Redis Connection Issues

```bash
kubectl -n virtuoso-api exec -it virtuoso-redis-0 -- redis-cli -a $REDIS_PASSWORD ping
```

### Ingress Issues

```bash
kubectl -n virtuoso-api describe ingress virtuoso-api-cli
kubectl -n ingress-nginx logs -l app.kubernetes.io/name=ingress-nginx
```

## Maintenance

### Rolling Updates

```bash
kubectl -n virtuoso-api set image deployment/virtuoso-api-cli api=virtuoso-api-cli:new-version
kubectl -n virtuoso-api rollout status deployment/virtuoso-api-cli
```

### Backup Redis Data

```bash
kubectl -n virtuoso-api exec virtuoso-redis-0 -- redis-cli -a $REDIS_PASSWORD BGSAVE
kubectl -n virtuoso-api cp virtuoso-redis-0:/data/dump.rdb ./redis-backup.rdb
```

## Clean Up

To remove all resources:

```bash
kubectl delete -k .
```
