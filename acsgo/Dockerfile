# ── Build stage ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Download dependencies first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /acsgo ./cmd/acsgo/main.go

# ── Runtime stage ───────────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /acsgo /acsgo
COPY --from=builder /build/web /web
COPY --from=builder /build/schema /schema

# Non-root user
USER 65534:65534

EXPOSE 7547 7548 9090

ENTRYPOINT ["/acsgo"]
