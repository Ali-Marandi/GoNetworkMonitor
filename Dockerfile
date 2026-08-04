# ---- Build Stage ----
FROM golang:1.22-bullseye AS builder

WORKDIR /app

# Install libpcap for CGO
RUN apt-get update && apt-get install -y libpcap-dev && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=2.0.0" \
    -o gonetmon ./cmd/gonetmon/

# ---- Runtime Stage ----
FROM debian:bullseye-slim

# Install libpcap runtime
RUN apt-get update && apt-get install -y \
    libpcap0.8 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/gonetmon .
COPY --from=builder /app/web ./web

# Create data directory
RUN mkdir -p /app/data

EXPOSE 8080

# Requires NET_ADMIN and NET_RAW capabilities for packet capture
# Run with: docker run --cap-add=NET_ADMIN --cap-add=NET_RAW --network=host ...
ENTRYPOINT ["/app/gonetmon"]
CMD ["--port", "8080"]
