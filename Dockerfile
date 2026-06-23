# ─── STAGE 1: COMPILE GO BINARY ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /workspace

# Copy dependencies first to leverage Docker caching layers
COPY go.mod go.sum ./
RUN go mod download

# Copy the restructured application source code blocks
COPY internal/ internal/
COPY cmd/ cmd/

# Build statically linked binary targeting the new entrypoint layout
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -a -o manager cmd/manager/main.go

# ─── STAGE 2: PRODUCTION RUNTIME ─────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /

# Bring over the compiled binary from the builder layer
COPY --from=builder /workspace/manager .

# 💡 FIX: Copy the entire internal/ai tree in one clean sweep from the builder
COPY --from=builder /workspace/internal/ai/ internal/ai/

# Run as a restricted, secure non-root user account
USER 65532:65532

ENTRYPOINT ["/manager"]