# VaultKit — Self-Hosted Secrets Manager
## Portfolio Project Plan

> **Dual-purpose portfolio project.** Demonstrates senior full-stack engineering (zero-knowledge encryption, auth, REST API design) *and* SRE/DevOps engineering (Kubernetes, Terraform, observability, secrets injection pipeline). One project, two audiences.

> **Note:** This document is the product vision and build plan. For how to actually
> set up, run, and develop the codebase, see the [README](../README.md). Where this
> plan and the README disagree on concrete versions (e.g. Go version), the README is
> the source of truth — this plan records the original intent.

---

## 0. Why VaultKit Exists

### The problem secrets management solves

Every software project produces secrets: API keys, database passwords, OAuth tokens, SSH keys, TLS certificates. The naive solution — storing them in a `.env` file, a shared Notion page, or worse, hardcoded in source code — is also the most common one. This creates a class of security failures that are entirely preventable yet happen constantly: leaked credentials in public repositories, shared Slack messages containing production passwords, ex-employees who still have access because no one rotated the credentials when they left.

Secrets management is the discipline of treating credentials as first-class infrastructure: stored securely, accessed with auditing, rotated on a schedule, and delivered to applications without human involvement.

### Why existing tools don't fully serve developers

The market has mature solutions at both ends of the spectrum but a gap in the middle:

**Consumer password managers** (1Password, Bitwarden, LastPass) are designed for individuals managing personal credentials through a browser extension. They have no concept of programmatic access, CI/CD integration, Kubernetes secrets injection, or per-secret audit logging. They are the right tool for a person's Netflix password, not for a team's production database.

**Enterprise secrets managers** (HashiCorp Vault, AWS Secrets Manager, CyberArk) solve the infrastructure problem well, but they come with significant operational cost. HashiCorp Vault requires a dedicated cluster, a careful unsealing ceremony, a storage backend, and ongoing maintenance. For a small team or an individual developer, standing up Vault is a multi-day project before you can store your first secret. AWS Secrets Manager solves the operational burden but creates cloud vendor lock-in and is opaque about what it does with your data.

**The gap:** a self-hosted, zero-knowledge secrets manager that a developer can deploy in an afternoon, that has first-class CLI and API access for pipelines, and that runs as a native Kubernetes workload without needing to trust a third party with unencrypted credentials.

### Why this is worth building as a portfolio project

Three reasons make this a better portfolio project than a generic CRUD app or a tutorial clone:

**1. It solves a real problem the builder actually has.** Developers genuinely need a place to store secrets that isn't a shared `.env` file. Building something you would actually use means the design decisions are grounded in reality, not invented constraints.

**2. It touches every layer of the stack in ways that matter to hiring managers.** Cryptographic architecture (SWE signal), a REST API with machine authentication (SWE + SRE signal), Kubernetes-native deployment with observability (SRE signal), and a CLI tool for pipeline use (SRE signal). No single tutorial covers all of this together.

**3. It mirrors a real production category.** HashiCorp Vault, Infisical, and Doppler are all funded companies solving this exact problem. Building a simplified version demonstrates you understand the domain — and gives you credible talking points in interviews about architectural tradeoffs: why zero-knowledge over server-side encryption, why AppRole over API keys, why PBKDF2 over bcrypt for key derivation.

### Why zero-knowledge specifically

The defining architectural decision in VaultKit is that the server is explicitly untrusted. This is not the default assumption in most web applications — typically the server is the trusted party and protects data on behalf of users.

Zero-knowledge flips this: the server stores only encrypted blobs it cannot read. This means a database breach, a rogue sysadmin, or a compromised cloud account cannot expose user secrets. The tradeoff is that there is no "forgot master password" recovery path — if the user loses their master password, the data is unrecoverable. This is a deliberate design choice, not an oversight, and explaining it in an interview demonstrates the ability to reason about security tradeoffs rather than just implementing features.

---

## 1. What Is VaultKit?

VaultKit is a self-hosted, zero-knowledge secrets manager built for developers and small teams. It allows users to store, organize, and access sensitive credentials — API keys, database passwords, tokens, SSH keys — through a web interface and a CLI tool.

"Zero-knowledge" means the server **never sees plaintext secrets**. All encryption and decryption happen on the client side. Even if the database is fully compromised, the secrets remain unreadable without the user's master password.

Unlike consumer password managers (1Password, Bitwarden), VaultKit is designed to be **deployed as infrastructure** — running inside a Kubernetes cluster, integrated with CI/CD pipelines, and surfaced as Prometheus metrics. This makes it relevant both as a product (SWE signal) and as a platform component (SRE signal).

---

## 2. How It Works — Core Concepts

### 2.1 Zero-Knowledge Encryption Model

```
User master password
        │
        ▼
  PBKDF2-SHA256            ← 600,000 iterations + random salt
  (key derivation)
        │
        ├──► Encryption key (AES-256-GCM)   ← used to encrypt vault entries
        │
        └──► Auth key                        ← used to authenticate with server
                                               (server stores hash of this, never the encryption key)
```

The master password is never sent to the server. The server receives only the auth key (itself derived and hashed), never the encryption key. Every vault entry is encrypted individually with a unique IV before transmission.

### 2.2 Secret Lifecycle

1. User creates a secret in the browser or CLI.
2. The frontend derives the encryption key from the master password (cached in memory for the session).
3. Secret is encrypted client-side with AES-256-GCM + random IV.
4. Encrypted blob is sent to the backend API via HTTPS.
5. Backend stores the encrypted blob in PostgreSQL — no plaintext ever written to disk.
6. To retrieve: encrypted blob is fetched, decrypted in the browser/CLI using the session encryption key.

### 2.3 Secret Injection for CI/CD (the SRE layer)

VaultKit exposes a machine-to-machine API for CI/CD pipelines and Kubernetes pods to fetch secrets at runtime:

```
CI/CD pipeline / K8s pod
        │
        ▼
  Authenticate via AppRole token (short-lived, scoped per project)
        │
        ▼
  VaultKit API returns encrypted secret
        │
        ▼
  Client library decrypts locally using project encryption key
        │
        ▼
  Secret available as environment variable or mounted file
```

This pattern mirrors how HashiCorp Vault Agent works — making VaultKit a recognisable, explainable architecture to any SRE interviewer.

---

## 3. User Scenarios

### Scenario A — Developer storing API keys (web UI)

Ana is a developer building a side project. She visits VaultKit, creates an account, and sets a master password. She adds entries: her AWS access key, a Stripe test secret, and a PostgreSQL connection string. Each entry has a name, value, tags, and optional notes. When she needs a credential, she logs in, the vault decrypts in her browser, and she copies the value. If she closes the tab, the decrypted key is gone from memory — the next login re-derives it.

### Scenario B — Team sharing project secrets (team vaults)

A three-person startup uses VaultKit for shared credentials. They create a team vault named `production`. Each team member has their own account, but they share the production vault encryption key — distributed securely out-of-band during onboarding. Any member can add, read, or rotate secrets in the shared vault. Access can be revoked by rotating the vault key.

### Scenario C — CI/CD pipeline fetching a secret (machine access)

A GitHub Actions workflow needs a database password for integration tests. The repository is configured with a VaultKit AppRole ID and a short-lived secret ID (rotated every 24 hours via a cron job). The workflow runs:

```bash
vaultkit auth --app-role $VAULTKIT_ROLE_ID --secret-id $VAULTKIT_SECRET_ID
export DB_PASSWORD=$(vaultkit secret get production/db/password)
```

VaultKit authenticates the pipeline, issues a scoped token valid for 60 minutes, and returns the encrypted secret. The CLI decrypts it locally. The plaintext password is only ever in the runner's memory.

### Scenario D — Kubernetes pod consuming a secret (K8s integration)

A Node.js microservice running in Kubernetes needs a third-party API key. Rather than storing it in a Kubernetes Secret (base64, not encrypted by default), the pod uses the VaultKit init container pattern:

1. An init container authenticates to VaultKit using the pod's Kubernetes service account.
2. It writes the decrypted secret to a shared in-memory volume (`tmpfs`).
3. The main container reads the secret from the volume as an environment variable.
4. When the pod terminates, the tmpfs volume is wiped.

### Scenario E — Secret rotation (ops scenario)

A team rotates their production database password. They update the value in VaultKit. A webhook fires to notify dependent services. The Kubernetes pods are rolled via a GitHub Actions workflow that triggers on vault updates. Prometheus metrics show the number of active secrets, last rotation timestamps, and failed access attempts — all visible in Grafana.

---

## 4. Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        User Layer                           │
│   Browser (React SPA)          CLI (Go)                     │
│   - Encrypts/decrypts          - Encrypts/decrypts          │
│   - Session key in memory      - Reads .vaultkit config     │
└────────────────────┬───────────────────┬────────────────────┘
                     │  HTTPS            │  HTTPS
┌────────────────────▼───────────────────▼────────────────────┐
│                    API Gateway / Ingress                     │
│              (NGINX Ingress Controller on K8s)              │
│              TLS termination via cert-manager               │
└──────────────────────────────┬──────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────┐
│                      Backend API                            │
│                    (Go / Kotlin)                            │
│   - Auth (JWT + refresh tokens)                             │
│   - Vault CRUD endpoints                                    │
│   - AppRole machine auth                                    │
│   - Webhook dispatch                                        │
│   - Prometheus /metrics endpoint                            │
└────────────┬──────────────────────────────┬─────────────────┘
             │                              │
┌────────────▼────────────┐   ┌────────────▼────────────────┐
│      PostgreSQL          │   │          Redis               │
│  (encrypted blobs,       │   │  (session tokens,            │
│   user accounts,         │   │   rate limiting,             │
│   audit log)             │   │   AppRole secret IDs)        │
└─────────────────────────┘   └─────────────────────────────┘
```

---

## 5. Component Details

### 5.1 Frontend (React + TypeScript)

**Purpose:** Provide the user-facing web vault UI. All cryptographic operations run in the browser — no secrets ever leave the client in plaintext.

**Tech stack:**
- React 18 + TypeScript
- Vite (build tool)
- TailwindCSS (styling)
- Web Crypto API (native browser AES-256-GCM, PBKDF2)
- React Query (server state management)
- Zustand (client state — session encryption key)

**Key screens:**
- Login / registration with master password setup
- Vault dashboard — list, search, filter by tag
- Secret detail — view (decrypted), edit, copy-to-clipboard with auto-clear
- Team vault management — invite, revoke access
- Audit log viewer — who accessed what and when
- Settings — 2FA setup (TOTP), session timeout config

**Security considerations:**
- Encryption key held only in Zustand memory store, never in localStorage or sessionStorage
- Auto-lock after configurable idle timeout (default: 15 minutes)
- Copy-to-clipboard clears after 30 seconds
- Content Security Policy headers enforced

**SWE signal:** Browser-side cryptography with Web Crypto API, proper key management, session security.

---

### 5.2 Backend API (Go or Kotlin)

**Language choice:** Go is recommended — aligns with the SRE toolchain and Kubernetes ecosystem. Kotlin is a valid alternative given your existing skill set.

**Purpose:** Store and retrieve encrypted blobs, handle authentication, manage machine tokens, emit metrics, dispatch webhooks.

**Tech stack (Go path):**
- Go 1.26+ (currently 1.26.4 — see the [README](../README.md) for the authoritative version)
- `chi` or `gin` (HTTP router)
- `pgx` (PostgreSQL driver)
- `go-redis` (Redis client)
- `golang-jwt/jwt` (JWT handling)
- `prometheus/client_golang` (metrics)
- `zap` (structured logging)

**Core API endpoints:**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Create account, derive auth key server-side hash |
| POST | `/api/auth/login` | Validate auth key hash, issue JWT + refresh token |
| POST | `/api/auth/refresh` | Rotate access token using refresh token |
| GET | `/api/vaults` | List user's vaults |
| POST | `/api/vaults` | Create vault |
| GET | `/api/vaults/:id/secrets` | List secrets (encrypted blobs only, no decryption) |
| POST | `/api/vaults/:id/secrets` | Store new encrypted secret |
| PUT | `/api/vaults/:id/secrets/:sid` | Update encrypted secret |
| DELETE | `/api/vaults/:id/secrets/:sid` | Delete secret (audit logged) |
| POST | `/api/approle/auth` | Machine authentication — returns scoped short-lived token |
| GET | `/api/approle/secret/:name` | Fetch encrypted secret by name (machine token required) |
| GET | `/metrics` | Prometheus metrics endpoint |

**Prometheus metrics exposed:**
- `vaultkit_secrets_total` — gauge, total secrets per vault
- `vaultkit_auth_attempts_total` — counter, labelled by success/failure
- `vaultkit_secret_access_total` — counter, labelled by vault and access method (web/cli/api)
- `vaultkit_approle_tokens_active` — gauge, active machine tokens
- `vaultkit_secret_age_days` — histogram, age distribution of secrets (rotation health)
- `vaultkit_api_request_duration_seconds` — histogram, API latency by endpoint

**SRE signal:** Custom Prometheus exporter built from scratch, structured JSON logging (Loki-ready), audit trail for incident investigation.

---

### 5.3 CLI Tool (Go)

**Purpose:** Allow developers and CI/CD pipelines to interact with VaultKit from the command line without a browser.

**Tech stack:** Go with `cobra` (CLI framework) + `keyring` (OS keychain integration for storing session tokens).

**Key commands:**

```bash
# Authentication
vaultkit login                         # interactive, prompts for master password
vaultkit auth --app-role ID --secret-id SECRET   # machine auth

# Secret operations
vaultkit secret get <vault>/<name>     # prints decrypted value to stdout
vaultkit secret set <vault>/<name>     # prompts for value, encrypts, stores
vaultkit secret list <vault>           # lists secret names (no values)
vaultkit secret rotate <vault>/<name>  # updates value, logs rotation event

# Team / vault management
vaultkit vault create <name>
vaultkit vault invite <email>

# Export (for migration)
vaultkit export --vault <name> --format env > .env.production
```

**SWE signal:** CLI tooling design, cobra patterns, OS keychain integration.
**SRE signal:** Designed for pipeline use — `secret get` exits with non-zero on failure, stdout-only output (pipeline-safe), JSON output mode (`--output json`).

---

### 5.4 Database (PostgreSQL)

**Schema (simplified):**

```sql
-- Users
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    auth_key_hash TEXT NOT NULL,       -- hash of derived auth key, never plaintext password
    kdf_salt    BYTEA NOT NULL,        -- salt for PBKDF2, stored per user
    kdf_iterations INT NOT NULL DEFAULT 600000,
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Vaults
CREATE TABLE vaults (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    owner_id    UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Secrets (all values are encrypted blobs)
CREATE TABLE secrets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id    UUID REFERENCES vaults(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    encrypted_value BYTEA NOT NULL,    -- AES-256-GCM ciphertext
    iv          BYTEA NOT NULL,        -- unique IV per entry
    tags        TEXT[],
    notes_encrypted BYTEA,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE(vault_id, name)
);

-- Audit log (append-only)
CREATE TABLE audit_events (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    UUID,                  -- user or app-role ID
    actor_type  TEXT NOT NULL,         -- 'user' | 'approle'
    action      TEXT NOT NULL,         -- 'secret.read' | 'secret.write' | 'secret.delete'
    vault_id    UUID,
    secret_id   UUID,
    ip_address  INET,
    created_at  TIMESTAMPTZ DEFAULT now()
);
```

**SWE signal:** Proper schema design, audit logging pattern, no plaintext storage (enforced at schema level — only BYTEA blobs).

---

### 5.5 Infrastructure & Deployment

#### Terraform (infrastructure provisioning)

Provisions the cloud environment from scratch with a single `terraform apply`. Target: AWS (EKS) or GCP (GKE).

**Resources provisioned:**
- VPC with public/private subnets
- Managed Kubernetes cluster (EKS or GKE)
- Managed PostgreSQL (RDS or Cloud SQL)
- Managed Redis (ElastiCache or Memorystore)
- S3 bucket for encrypted backups
- IAM roles for pod-level cloud access (IRSA on AWS)
- ACM/GCP-managed TLS certificates

**Terraform structure:**
```
terraform/
├── modules/
│   ├── vpc/
│   ├── eks/
│   ├── rds/
│   └── redis/
├── environments/
│   ├── dev/
│   └── prod/
└── main.tf
```

#### Kubernetes manifests / Helm chart

VaultKit ships as a Helm chart for easy self-hosting by others.

**Key Kubernetes resources:**
- `Deployment` — backend API (3 replicas in prod)
- `Deployment` — frontend (served via NGINX, 2 replicas)
- `HorizontalPodAutoscaler` — scales backend on CPU + custom `vaultkit_api_request_duration_seconds` metric
- `PodDisruptionBudget` — ensures at least 2 backend pods always available
- `Ingress` — NGINX ingress with cert-manager for TLS
- `ServiceMonitor` — tells Prometheus Operator to scrape the backend `/metrics`
- `CronJob` — rotates AppRole secret IDs every 24 hours
- `NetworkPolicy` — restricts pod-to-pod traffic (only ingress → backend → postgres/redis)

**Secrets in Kubernetes:**
Kubernetes Secrets (for DB credentials, JWT signing key) are stored as `ExternalSecret` resources, synced from AWS Secrets Manager / GCP Secret Manager via the External Secrets Operator. No plaintext credentials in the Git repository.

#### CI/CD Pipeline (GitHub Actions)

```
Push to main
    │
    ├── Lint + unit tests (Go / TypeScript)
    ├── Security scan (gosec, npm audit, truffleHog for leaked secrets)
    ├── Build Docker images (multi-stage, distroless final image)
    ├── Push to container registry (GHCR or ECR)
    ├── Terraform plan (preview infra changes)
    └── Helm upgrade (rolling deploy to K8s)
             │
             └── Post-deploy health check
                      │
                      ├── Pass → done
                      └── Fail → automatic rollback (helm rollback)
```

**SRE signal:** Automated rollback on failed health check, secret scanning in CI, distroless images (minimal attack surface), HPA on custom metric.

---

### 5.6 Observability Stack

**Components:**
- **Prometheus** — scrapes `/metrics` from backend via ServiceMonitor
- **Grafana** — dashboards for API health, secret activity, auth patterns
- **Loki** — aggregates structured JSON logs from all pods
- **Alertmanager** — fires alerts on anomalous patterns

**Key Grafana dashboards:**
1. **API health** — request rate, error rate, p50/p99 latency per endpoint
2. **Vault activity** — secret reads/writes per vault over time, rotation age heatmap
3. **Auth security** — failed login attempts (rate alerts for brute force), AppRole token lifecycle
4. **Infrastructure** — pod CPU/memory, HPA scaling events, PVC usage

**Key Alertmanager rules:**
- `HighAuthFailureRate` — more than 20 failed auth attempts in 5 minutes from the same IP
- `SecretNotRotated` — any secret older than 90 days (fires to Slack)
- `APILatencyHigh` — p99 latency > 500ms for 5 minutes
- `PodRestartLoop` — backend pod restarting more than 3 times in 10 minutes

**SRE signal:** SLO defined as 99.9% uptime + p99 < 200ms; error budget tracked in Grafana; Alertmanager rules with actionable thresholds.

---

## 6. What Each Layer Signals to Interviewers

### SWE Hiring Manager sees:
- Cryptographic architecture (AES-256-GCM, PBKDF2, zero-knowledge model)
- Secure API design (auth key hashing, JWT lifecycle, AppRole machine auth)
- Full-stack depth (React + Web Crypto, Go/Kotlin backend, PostgreSQL schema)
- Audit logging and compliance-conscious data design
- CLI tooling with production-quality UX

### SRE Hiring Manager sees:
- Custom Prometheus exporter written from scratch in Go
- Kubernetes deployment with HPA on a custom metric
- External Secrets Operator — no credentials in Git, ever
- Terraform-provisioned infrastructure, destroy/recreate in one command
- Observability: metrics + logs + alerts, with defined SLOs
- CI/CD with automated rollback
- Security-conscious operations: NetworkPolicy, distroless images, AppRole rotation

---

## 7. Build Sequence

Build in this order — each phase produces something usable before moving to the next.

> **Deploy-first amendment.** The original sequence put Kubernetes in Phase 3. In
> practice we front-loaded a thin **Phase 0** that proves the full
> source → image → CI/CD → pod path while the app is still trivial, so feature work
> ships continuously instead of hitting deployment surprises late. The heavyweight
> infra (Helm, Terraform, Ingress, cert-manager, External Secrets) stays in Phase 3.

**Phase 0 — Deployable walking skeleton (in progress)**
Make the scaffold deployable end-to-end with mocked data and minimal instrumentation.
- [x] App made K8s-correct: SIGTERM graceful shutdown, `/healthz` probe, env-driven `PORT`
- [x] Multi-stage distroless Dockerfile (static binary, nonroot, ~10MB) + `.dockerignore`
- [ ] CI/CD: build & publish image to GHCR on merge to `main`
- [ ] Minimal K8s manifests (`deploy/k8s/`): Deployment + Service with probes

**Phase 1 — Core app (4–6 weeks)**
Get the zero-knowledge vault working locally. Docker Compose for local dev (backend + postgres + redis). No Kubernetes yet.
- [ ] PostgreSQL schema + migrations
- [ ] Go backend: auth endpoints + vault CRUD
- [ ] Web Crypto encryption layer (TypeScript utility)
- [ ] React frontend: login, vault list, secret CRUD
- [ ] CLI: `login`, `secret get`, `secret set`
- [ ] Unit tests for crypto layer (encrypt → store → decrypt round-trip)

**Phase 2 — SRE instrumentation (2–3 weeks)**
Add the observability and machine-access layers.
- [ ] Prometheus metrics endpoint on backend
- [ ] AppRole auth + scoped tokens
- [ ] Structured JSON logging (zap)
- [ ] Local Prometheus + Grafana via Docker Compose
- [ ] Alertmanager rules (local)

**Phase 3 — Production infrastructure (3–4 weeks)**
Move from Docker Compose to real Kubernetes.
- [ ] Helm chart for VaultKit
- [ ] Terraform: VPC + EKS/GKE + RDS + Redis
- [ ] cert-manager + NGINX ingress
- [ ] External Secrets Operator (sync DB creds from cloud secrets manager)
- [ ] HPA + PodDisruptionBudget
- [ ] GitHub Actions CI/CD with rollback

**Phase 4 — Polish for portfolio (1–2 weeks)**
Make it presentable and tell the story.
- [ ] Architecture diagram (draw.io or Mermaid in README)
- [ ] README: what it is, how to self-host it, design decisions
- [ ] Blog post or LinkedIn article: "Why I built a zero-knowledge secrets manager" — explain the crypto decisions and the K8s deployment
- [ ] Record a 3-minute demo video

---

## 8. Technology Decisions Summary

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Backend language | Go | Aligns with K8s ecosystem; Prometheus client is first-class; strong for CLI tooling |
| Frontend | React + TypeScript | Your existing skill set; Web Crypto API is well-supported |
| Encryption | AES-256-GCM + PBKDF2 | Industry standard; same as Bitwarden and 1Password |
| Database | PostgreSQL | BYTEA support; strong audit logging; managed options on all clouds |
| Cache / tokens | Redis | TTL-native for AppRole secret ID expiry; fast rate limiting |
| Container | Kubernetes + Helm | Portable; shows real-world deployment knowledge |
| IaC | Terraform | Most in-demand IaC tool in SRE job postings |
| Observability | Prometheus + Grafana + Loki | Open-source standard; directly maps to SRE cert requirements |
| CI/CD | GitHub Actions | Native to GitHub; widely used; easy to showcase |

---

*This document is a living plan. Scope phases 3 and 4 based on time available — phase 1 + 2 alone produce a strong SWE portfolio piece; all four phases together produce a strong SRE portfolio piece.*
