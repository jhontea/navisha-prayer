# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git (needed by some go modules) and CA certificates
RUN apk add --no-cache git ca-certificates tzdata

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary.
# CGO is disabled — this project uses the pure-Go modernc/glebarez SQLite driver,
# so no C toolchain is required.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20

WORKDIR /app

# CA certs (for HTTPS calls to Aladhan / Open-Meteo) and timezone data
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

# Copy the compiled binary
COPY --from=builder /app/server /app/server

# Persisted SQLite data lives here (mount a volume in production)
RUN mkdir -p /app/data && chown -R app:app /app

USER app

EXPOSE 8010

# Healthcheck hits the API health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8010/api/v1/health || exit 1

ENTRYPOINT ["/app/server"]
