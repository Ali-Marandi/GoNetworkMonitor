# ---- Build Stage ----
FROM golang:1.25-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y libpcap-dev && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=3.1.0" \
    -o gonetmon ./cmd/gonetmon/

# ---- Runtime Stage ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libpcap0.8 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/gonetmon .

RUN mkdir -p /app/data

EXPOSE 8080

# Requires NET_ADMIN and NET_RAW capabilities for packet capture.
ENTRYPOINT ["/app/gonetmon"]
CMD ["--port", "8080"]
