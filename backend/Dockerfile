# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install ca-certificates and git
RUN apk add --no-cache ca-certificates git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Final production stage
FROM alpine:3.20

WORKDIR /app

# Install root CA certificates for TLS (MongoDB Atlas, Redis TLS)
RUN apk add --no-cache ca-certificates tzdata

# Copy compiled binary from builder
COPY --from=builder /app/server /app/server

# Render provides the PORT env var dynamically
ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

CMD ["/app/server"]
