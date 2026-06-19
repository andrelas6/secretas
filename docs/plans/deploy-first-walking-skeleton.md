# Plan: Deploy-first walking skeleton for VaultKit

## Context

VaultKit is currently an early Phase-1 scaffold: a Go HTTP server with one
`/secret` echo endpoint, no DB, no auth, no persistence. The goal is to
**deploy to a Kubernetes cluster now** — deliberately ahead of the
`PROJECT_PLAN.md` phase ordering (K8s was Phase 3).

The rationale is **deploy-first / walking-skeleton** development: prove the full
path — source → container image → CI/CD → running pod — while the app is trivial,
so later feature work ships continuously instead of discovering deployment problems
after a large codebase is written. We lay the foundation (deployable + CI/CD) first;
mocked data and minimal instrumentation are fine for now.

This plan makes the existing app genuinely container/K8s-correct, containerizes it,
publishes the image via CI, and provides minimal manifests to deploy it — then
records the strategy change in the project docs.

## Decisions already locked
- **OTel: leave as-is** (writes telemetry to `.otel/` under the working dir at
  runtime). Implication: the container needs a writable working dir; we do **not**
  enable `readOnlyRootFilesystem` yet. Hardening that comes naturally later when OTel
  is switched to stdout.

## Changes

### 1. App: make it actually K8s-correct (`cmd/vaultkitd/main.go`)
Small, targeted edits — these are real gaps, not nice-to-haves:

- **SIGTERM handling (real bug).** `main.go:32` uses
  `signal.NotifyContext(ctx, os.Interrupt)`, catching only SIGINT (Ctrl-C).
  Kubernetes sends **SIGTERM** on pod termination, so graceful shutdown never fires
  today — the pod gets SIGKILLed after the grace period. Add `syscall.SIGTERM`:
  `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`.
- **Health endpoint for probes.** Add a `/healthz` handler returning `200 OK` in
  `newHttpHandler()` (alongside the existing `/secret` mux registration). K8s
  liveness/readiness probes need a cheap endpoint. One combined endpoint is fine for
  the skeleton; split into `/livez` + `/readyz` later when there's a DB to gate
  readiness on.
- **Env-driven port.** Replace hardcoded `Addr: ":3001"` (`main.go:49`) with a
  `PORT` env var defaulting to `3001`. Keeps the dev experience identical, lets the
  manifest/Service set the port declaratively.

### 2. Dockerfile + `.dockerignore` (new, repo root)
Multi-stage, distroless — matches `PROJECT_PLAN.md` §5.5 ("multi-stage, distroless
final image"):

- **Builder stage:** `golang:1.26.4` (matches `go.mod`). Build static binary:
  `CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /vaultkitd ./cmd/vaultkitd`.
  Copy `go.mod`/`go.sum` and `go mod download` first for layer caching.
- **Final stage:** `gcr.io/distroless/static-debian12:nonroot` (runs as uid 65532).
  Copy the binary. **Pre-create a writable `/app/.otel` dir owned by 65532** in the
  builder and `COPY --chown=65532:65532` it over (distroless has no shell, so the dir
  must be created before copy), set `WORKDIR /app`, `EXPOSE 3001`,
  `ENTRYPOINT ["/vaultkitd"]`. The pre-created writable dir is what lets the
  leave-as-is OTel file writes succeed.
- **`.dockerignore`:** exclude `.git`, `bin/`, the committed `vaultkitd` binary,
  `.otel/`, `docs/`, `*.md`, `.githooks/` — keep build context minimal.

### 3. CI/CD: build & publish image (`.github/workflows/docker-publish.yml`, new)
Separate workflow from `ci.yml` (keeps the fast test loop independent):

- Trigger on push to `main` (and optionally tags).
- `permissions: packages: write`, log in to **GHCR** with the built-in
  `GITHUB_TOKEN`, build with `docker/build-push-action`, push to
  `ghcr.io/andrelas6/secretas` tagged with both the commit SHA and `latest`.
- Gate on the existing CI passing (or at minimum run after merge to `main`).

### 4. Minimal K8s manifests (`deploy/k8s/`, new)
Raw YAML (Helm stays deferred to Phase 3 per the plan):

- `deployment.yaml` — 1 replica, image `ghcr.io/andrelas6/secretas:latest`,
  `containerPort 3001`, liveness + readiness probes hitting `/healthz`, modest
  resource requests/limits, `securityContext: runAsNonRoot: true` (matches distroless
  nonroot). **No `readOnlyRootFilesystem`** yet (OTel writes files).
- `service.yaml` — `ClusterIP` exposing port 3001.
- Brief `deploy/k8s/README.md` with `kubectl apply -k` / `port-forward` usage.

### 5. Docs: record the strategy change
- **`docs/PROJECT_PLAN.md` §7:** insert a **"Phase 0 — Deployable walking skeleton"**
  before Phase 1, and add a short note explaining the deploy-first reordering (Docker
  + CI/CD + minimal manifests now; Helm/Terraform/ingress remain Phase 3).
- **`AGENTS.md` "Project state":** update Phase + next concrete step to reflect
  deploy-first foundation.
- **`README.md` status + project layout:** note the Dockerfile, `deploy/k8s/`, and how
  to build/run the container.

## Files
- Edit: `cmd/vaultkitd/main.go`
- New: `Dockerfile`, `.dockerignore`
- New: `.github/workflows/docker-publish.yml`
- New: `deploy/k8s/deployment.yaml`, `deploy/k8s/service.yaml`, `deploy/k8s/README.md`
- Edit: `docs/PROJECT_PLAN.md`, `AGENTS.md`, `README.md`

## Verification
1. **Build/tests:** `go build ./...` and `go test -race ./...` pass (hooks/CI parity).
2. **Local container:**
   - `docker build -t vaultkit:dev .`
   - `docker run --rm -p 3001:3001 vaultkit:dev`
   - `curl -s localhost:3001/healthz` → `200`
   - `curl -s localhost:3001/secret -d '{"name":"x","encrypted_value":"y","iv":"z","tags":[],"notes_encrypted":""}'` → echo response.
   - **Graceful shutdown:** `docker stop` (sends SIGTERM) → logs show the
     "server shutting down..." path, container exits 0 (not killed after timeout).
3. **CI:** push branch/PR → existing `ci.yml` green; merge to `main` →
   `docker-publish.yml` pushes `ghcr.io/andrelas6/secretas:<sha>` and `:latest`.
4. **Cluster:** `kubectl apply -k deploy/k8s/` (or kind), `kubectl get pods` Ready,
   `kubectl port-forward svc/vaultkit 3001:3001`, then re-run the curl checks.

## Out of scope (deferred, by design)
- Helm chart, Terraform, NGINX ingress, cert-manager, External Secrets (Phase 3).
- Switching OTel to stdout + enabling `readOnlyRootFilesystem` (natural next hardening
  step once telemetry no longer writes files).
- Real DB / auth / persistence (Phase 1 feature work, unblocked by this foundation).
