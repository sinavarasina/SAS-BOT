# Multi-stage build untuk optimasi ukuran image

# Stage 1: Build
FROM golang:1.24 AS builder

WORKDIR /app

# Install dependencies (using apt for Debian-based golang image)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sas-bot ./cmd/sas-bot/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install ca-certificates untuk HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Copy binary dari builder
COPY --from=builder /app/sas-bot .

# Copy config files dan JSON data
COPY json/ ./json/
COPY .env .env
COPY sas-sindang-anom-58dc655332ab.json .

# Create uploads directory
RUN mkdir -p uploads

# Expose port (optional, untuk reference)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD test -f /app/sas-bot || exit 1

# Run aplikasi
CMD ["./sas-bot"]
