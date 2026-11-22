# Build stage
FROM golang:1.27-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags=-static" \
    -o /app/bin/kicaugo ./cmd/api

# Runtime stage - distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /app/bin/kicaugo /kicaugo

# Copy migrations for runtime execution (from internal/database/migrations where embed.go reads them)
COPY --from=builder /app/internal/database/migrations /migrations

# Non-root user (UID 65532 = nonroot in distroless)
USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/kicaugo"]