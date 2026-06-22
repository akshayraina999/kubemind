# ─── STAGE 1: COMPILE GO BINARY ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copy dependencies first to leverage Docker caching layers
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the core application source code
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build statically linked binary with optimization flags (-s -w strips debug data)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -a -o manager cmd/main.go

# ─── STAGE 2: PRODUCTION RUNTIME ─────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /

# Bring over the compiled binary from the builder layer
COPY --from=builder /workspace/manager .

# Bring over your system prompt matrix text asset
COPY ai/prompts/remediation_prompt.txt ai/prompts/remediation_prompt.txt

# Run as a restricted, secure non-root user account (IDs match standard guidelines)
USER 65532:65532

ENTRYPOINT ["/manager"]