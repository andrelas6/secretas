# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.26.4 AS builder

WORKDIR /src

# Cache deps first — this layer only re-runs when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Build a fully static binary (CGO off) so it runs on distroless/static.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/vaultkitd ./cmd/vaultkitd

# Pre-create the runtime working dir owned by the distroless nonroot uid (65532).
# The app writes OTel telemetry to ./.otel at runtime, and distroless has no shell
# to mkdir/chown in the final stage — so it must exist before we COPY it over.
RUN mkdir -p /out/app/.otel

# ---- Runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot

# Writable working dir (owned by nonroot) for the OTel file exporters.
COPY --from=builder --chown=65532:65532 /out/app /app
COPY --from=builder /out/vaultkitd /app/vaultkitd

WORKDIR /app
USER 65532:65532

EXPOSE 3001
ENTRYPOINT ["/app/vaultkitd"]
