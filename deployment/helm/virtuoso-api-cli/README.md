# Virtuoso API CLI Helm Chart

A production-ready Helm chart for deploying the Virtuoso API CLI - an AI-friendly interface for Virtuoso's test automation platform.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- PV provisioner support in the underlying infrastructure (if Redis persistence is enabled)
- Ingress controller (if ingress is enabled)
- Prometheus operator (if ServiceMonitor is enabled)
- External Secrets Operator (if external secrets are enabled)

## Installing the Chart

### Add the repository (if published)

```bash
helm repo add virtuoso https://charts.virtuoso.qa
helm repo update
```

### Install with default values

```bash
helm install my-virtuoso-api virtuoso/virtuoso-api-cli \
  --namespace virtuoso-api \
  --create-namespace \
  --set secret.apiToken="your-api-token" \
  --set config.organization.id="your-org-id"
```

### Install from local directory

```bash
helm install my-virtuoso-api ./virtuoso-api-cli \
  --namespace virtuoso-api \
  --create-namespace \
  --set secret.apiToken="your-api-token" \
  --set config.organization.id="your-org-id"
```

### Install with environment-specific values

```bash
# Development
helm install virtuoso-api-dev ./virtuoso-api-cli \
  -f values.yaml \
  -f values.dev.yaml \
  --namespace virtuoso-dev \
  --create-namespace \
  --set secret.apiToken="dev-api-token" \
  --set config.organization.id="2242-dev"

# Staging
helm install virtuoso-api-staging ./virtuoso-api-cli \
  -f values.yaml \
  -f values.staging.yaml \
  --namespace virtuoso-staging \
  --create-namespace \
  --set secret.apiToken="staging-api-token" \
  --set config.organization.id="2242-staging"

# Production
helm install virtuoso-api-prod ./virtuoso-api-cli \
  -f values.yaml \
  -f values.production.yaml \
  --namespace virtuoso-production \
  --create-namespace
```

## Uninstalling the Chart

```bash
helm uninstall my-virtuoso-api -n virtuoso-api
```

## Configuration

### Required Values

| Parameter                | Description                       | Default                                       |
| ------------------------ | --------------------------------- | --------------------------------------------- |
| `secret.apiToken`        | Virtuoso API authentication token | `""` (required if not using external secrets) |
| `config.organization.id` | Virtuoso organization ID          | `""` (required)                               |

### Common Configuration Options

| Parameter                   | Description             | Default                      |
| --------------------------- | ----------------------- | ---------------------------- |
| `replicaCount`              | Number of replicas      | `3`                          |
| `image.repository`          | Image repository        | `virtuoso-api-cli`           |
| `image.tag`                 | Image tag               | `""` (uses chart appVersion) |
| `image.pullPolicy`          | Image pull policy       | `IfNotPresent`               |
| `service.type`              | Kubernetes service type | `ClusterIP`                  |
| `service.port`              | Service port            | `80`                         |
| `ingress.enabled`           | Enable ingress          | `false`                      |
| `ingress.hosts[0].host`     | Ingress hostname        | `api-cli.virtuoso.qa`        |
| `resources.limits.cpu`      | CPU limit               | `1000m`                      |
| `resources.limits.memory`   | Memory limit            | `2Gi`                        |
| `resources.requests.cpu`    | CPU request             | `500m`                       |
| `resources.requests.memory` | Memory request          | `1Gi`                        |

### Advanced Configuration

#### External Secrets

To use external secrets (e.g., from Vault, AWS Secrets Manager):

```yaml
externalSecrets:
  enabled: true
  secretStoreRef:
    name: vault-backend
    kind: ClusterSecretStore
  data:
    - secretKey: api-token
      remoteRef:
        key: virtuoso/api-cli
        property: token
```

#### Redis Configuration

Redis is enabled by default for session management. To use an external Redis:

```yaml
redis:
  enabled: false

config:
  redis:
    host: external-redis.example.com
    port: 6379
```

#### Autoscaling

Enable horizontal pod autoscaling:

```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

#### Monitoring

Enable Prometheus ServiceMonitor:

```yaml
serviceMonitor:
  enabled: true
  labels:
    prometheus: kube-prometheus
```

### Full Values Reference

See [values.yaml](values.yaml) for a complete list of configurable parameters.

## Security Considerations

1. **API Token**: Never commit API tokens to version control. Use external secrets or sealed secrets.

2. **Network Policies**: Consider implementing network policies to restrict traffic:

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: virtuoso-api-cli-netpol
     namespace: virtuoso-api
   spec:
     podSelector:
       matchLabels:
         app.kubernetes.io/name: virtuoso-api-cli
     policyTypes:
     - Ingress
     - Egress
     ingress:
     - from:
       - namespaceSelector:
           matchLabels:
             name: ingress-nginx
       ports:
       - protocol: TCP
         port: 8000
     egress:
     - to:
       - namespaceSelector: {}
       ports:
       - protocol: TCP
         port: 443
     - to:
       - podSelector:
           matchLabels:
             app.kubernetes.io/name: virtuoso-api-cli
             app.kubernetes.io/component: redis
       ports:
       - protocol: TCP
         port: 6379
   EOF
   ```

3. **Pod Security Standards**: The chart follows security best practices:
   - Runs as non-root user (UID 1000)
   - Read-only root filesystem
   - No privilege escalation
   - Drops all capabilities

## Backup and Recovery

### Redis Backup (if enabled)

```bash
# Create backup
kubectl exec -n virtuoso-api virtuoso-api-cli-redis-0 -- redis-cli -a $REDIS_PASSWORD --rdb /data/backup.rdb

# Copy backup locally
kubectl cp virtuoso-api/virtuoso-api-cli-redis-0:/data/backup.rdb ./redis-backup.rdb
```

## Troubleshooting

### Check deployment status

```bash
kubectl get all -n virtuoso-api -l app.kubernetes.io/name=virtuoso-api-cli
```

### View logs

```bash
# API logs
kubectl logs -n virtuoso-api -l app.kubernetes.io/name=virtuoso-api-cli -f

# Redis logs (if enabled)
kubectl logs -n virtuoso-api -l app.kubernetes.io/component=redis -f
```

### Debug pod

```bash
kubectl exec -it -n virtuoso-api deployment/virtuoso-api-cli -- /bin/sh
```

### Common Issues

1. **Pod not starting**: Check resource limits and node capacity
2. **Auth errors**: Verify API token is correctly set
3. **Redis connection errors**: Check Redis password and network connectivity
4. **Ingress not working**: Verify ingress controller is installed and configured

## Upgrading

### Minor version upgrades

```bash
helm upgrade my-virtuoso-api ./virtuoso-api-cli \
  --namespace virtuoso-api \
  --reuse-values
```

### Major version upgrades

1. Review the changelog and breaking changes
2. Backup any persistent data
3. Update values file as needed
4. Perform the upgrade:
   ```bash
   helm upgrade my-virtuoso-api ./virtuoso-api-cli \
     --namespace virtuoso-api \
     -f values.yaml \
     -f values.production.yaml
   ```

## Development

### Testing the chart

```bash
# Lint the chart
helm lint ./virtuoso-api-cli

# Dry run installation
helm install my-test ./virtuoso-api-cli \
  --debug \
  --dry-run \
  --set secret.apiToken="test" \
  --set config.organization.id="test"

# Template rendering
helm template my-test ./virtuoso-api-cli \
  --set secret.apiToken="test" \
  --set config.organization.id="test"
```

### Package the chart

```bash
helm package ./virtuoso-api-cli
```

## Support

For issues and questions:

- GitHub Issues: [virtuoso/virtuoso-api-cli](https://github.com/virtuoso/virtuoso-api-cli)
- Email: support@virtuoso.qa

## License

This Helm chart is licensed under the Apache License 2.0. See [LICENSE](../../LICENSE) for details.
