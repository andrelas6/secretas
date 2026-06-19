# VaultKit

A self-hosted, zero-knowledge secrets manager. The server stores only encrypted
blobs it can never read — all encryption/decryption happens client-side.

This README is the **developer setup guide**: how to run, build, and work on the
codebase. For the product vision, architecture, and full build plan, see
[`docs/PROJECT_PLAN.md`](docs/PROJECT_PLAN.md).

**Status:** Phase 0 (deploy-first walking skeleton). A Go HTTP server with a placeholder
`/secret` endpoint and a `/healthz` probe, containerized (distroless) and deployable to
Kubernetes (`deploy/k8s/`). PostgreSQL, auth, and the vault model are not built yet.

---

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | **1.26.4** | Matches the `go` directive in `go.mod`. |
| golangci-lint | **v2.12.2** | Required for the pre-commit hook and CI. Must be a build compiled with Go 1.26+ — see [golangci-lint](#golangci-lint) below. |

Clone and pull dependencies:

```bash
git clone git@github.com:andrelas6/secretas.git
cd secretas
go mod download
```

Then enable the git hooks (one-time, see [Git hooks](#git-hooks)):

```bash
git config core.hooksPath .githooks
```

---

## Running for development

Run the server directly from source — no build step, recompiles on each run:

```bash
go run ./cmd/vaultkitd
```

The server listens on **`:3001`**. Stop it with `Ctrl-C` (it shuts down gracefully).

## Building and running the binary

Compile a standalone binary (output goes to `bin/`, which is gitignored):

```bash
go build -o bin/vaultkitd ./cmd/vaultkitd
./bin/vaultkitd
```

Same result as `go run`, but you get a single distributable binary and a faster start.

## Running in a container

Build the image and run it (multi-stage, distroless, runs as nonroot):

```bash
docker build -t vaultkit:dev .
docker run --rm -p 3001:3001 vaultkit:dev
```

`PORT` is read from the environment (default `3001`); override with
`docker run -e PORT=8080 -p 8080:8080 vaultkit:dev`. The server catches `SIGTERM`,
so `docker stop` shuts it down gracefully.

## Trying the API

The only endpoint so far is `POST /secret`. It currently **echoes back** the secret
payload you send — it decodes the JSON into a `SecretDTO`, rejects any unknown fields,
and returns the same object. Nothing is persisted yet (no database). It's a scaffold
for the real "store encrypted secret" endpoint.

With the server running, send a sample request:

```bash
curl -s http://localhost:3001/secret \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "stripe-api-key",
    "encrypted_value": "9b2c...base64-ciphertext...==",
    "iv": "f1a0...base64-iv...==",
    "tags": ["payments", "prod"],
    "notes_encrypted": "a3e4...base64-notes...=="
  }'
```

Response (the decoded payload, echoed back):

```json
{"name":"stripe-api-key","encrypted_value":"9b2c...base64-ciphertext...==","iv":"f1a0...base64-iv...==","tags":["payments","prod"],"notes_encrypted":"a3e4...base64-notes...=="}
```

Notes on current behavior:
- **Fields** come from `SecretDTO` (`internal/secret/controller/secret_dto.go`):
  `name`, `encrypted_value`, `iv`, `tags`, `notes_encrypted`.
- **Unknown fields are rejected** — sending a field not in the DTO returns
  `400 Invalid json` (the decoder uses `DisallowUnknownFields`).
- Per the zero-knowledge model, `encrypted_value`/`iv`/`notes_encrypted` are meant to
  be **client-side ciphertext** — the server never receives plaintext. The values
  above are illustrative placeholders.

---

## Project layout

```
cmd/vaultkitd/        Server entrypoint (main, graceful shutdown, OTel setup)
internal/
  env/                Minimal .env loader with OS-env fallback (PORT, ...)
  k8s/health/         /healthz handler for liveness/readiness probes
  observability/      OpenTelemetry tracing setup
  secret/controller/  HTTP handler + DTO for the /secret endpoint
migrations/           (placeholder) SQL migrations — not built yet
Dockerfile            Multi-stage, distroless container image
deploy/k8s/           Minimal Kubernetes manifests (Deployment + Service)
.githooks/            pre-commit / pre-push hooks (see below)
docs/PROJECT_PLAN.md  Product vision, architecture, build plan
```

---

## Git hooks

The repo ships git hooks in `.githooks/` that mirror CI, catching issues before they
reach a PR. They are version-controlled but **inactive until you opt in per clone**:

```bash
git config core.hooksPath .githooks
```

| Hook | Runs | When |
|------|------|------|
| `pre-commit` | `gofmt` check → `go vet ./...` → `golangci-lint run` | every commit (fast static checks) |
| `pre-push` | `go test -race ./...` | every push (slower, race-enabled tests) |

If a hook fails, the commit/push is aborted with a message describing what to fix.

**Bypassing** (for WIP commits you don't intend to push): add `--no-verify`, e.g.
`git commit --no-verify`. CI still enforces everything, so nothing un-checked reaches
`main`.

---

## golangci-lint

The pre-commit hook and CI both run **golangci-lint v2.12.2**. Install the **binary** —
the project's [official install docs](https://golangci-lint.run/docs/welcome/install/local/)
recommend this over `go install` (reproducible, no dependency contamination):

```bash
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2
```

Verify it's a **Go 1.26+ build**:

```bash
golangci-lint version
# golangci-lint has version 2.12.2 built with go1.26.x ...
```

> **Why the Go version matters:** golangci-lint refuses to analyze a module whose `go`
> directive (here, 1.26) is newer than the Go version the linter binary was *built
> with*. An older binary (e.g. built with go1.25) fails with
> *"the Go language version used to build golangci-lint is lower than the targeted Go
> version"*. v2.12.2's released binary is built with go1.26.2, so it works.

CI pins this **same version** in [`.github/workflows/ci.yml`](.github/workflows/ci.yml)
so local and CI run an identical linter set. **When bumping, change it in both places**
(this README's prereq table + the workflow).

---

## Common commands

```bash
go run ./cmd/vaultkitd        # run the server from source
go build -o bin/vaultkitd ./cmd/vaultkitd   # build a binary
go test ./...                 # run tests
go test -race ./...           # run tests with the race detector (as pre-push does)
go vet ./...                  # vet
gofmt -w .                    # format all files in place
golangci-lint run             # lint (same as pre-commit / CI)
```

---

## Continuous integration

GitHub Actions (`.github/workflows/ci.yml`) runs on every pull request and on pushes to
`main`. It performs: dependency verification (`go mod verify`), a `go mod tidy` drift
check, `gofmt`, `go vet`, `golangci-lint`, `go build`, and `go test -race -cover`. A
merge to `main` is a push, so the same checks run post-merge.
