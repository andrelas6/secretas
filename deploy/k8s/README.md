# VaultKit — Kubernetes manifests

Minimal manifests to run VaultKit as a pod. This is the deploy-first walking
skeleton: a `Deployment` + `Service` with health probes and a hardened security
context. Helm, Ingress, TLS, and autoscaling are deferred to Phase 3 (see
`docs/PROJECT_PLAN.md`).

## Contents

| File | Purpose |
|------|---------|
| `deployment.yaml` | 1-replica Deployment, `/healthz` liveness+readiness probes, nonroot security context, resource requests/limits |
| `service.yaml` | `ClusterIP` Service exposing port 3001 |
| `kustomization.yaml` | Bundles the two for `kubectl apply -k` |

## Deploy

```bash
kubectl apply -k deploy/k8s/
kubectl rollout status deploy/vaultkit
```

## Verify

```bash
kubectl get pods -l app=vaultkit
kubectl port-forward svc/vaultkit 3001:3001
# in another shell:
curl -s localhost:3001/healthz   # -> 200
```

## Gotchas

- **Image visibility.** `docker-publish.yml` pushes to `ghcr.io/andrelas6/secretas`.
  GHCR packages are **private by default** — either make the package public, or add
  an `imagePullSecret` to the pod so the cluster can pull it.
- **`:latest` tag.** Fine for a skeleton, but for real rollouts pin the
  `sha-<commit>` tag the publish workflow produces — it's immutable and traceable.
- **No `readOnlyRootFilesystem`.** Deliberately omitted: the app writes OTel files
  to `./.otel`. It can be enabled once telemetry moves to stdout.
