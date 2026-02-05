# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o inspection-tool ./cmd

# Final stage
FROM alpine:latest

WORKDIR /app

# Install required packages
RUN apk add --no-cache ca-certificates openssh-client

# Copy binary from builder
COPY --from=builder /build/inspection-tool /app/inspection-tool

# Create reports directory
RUN mkdir -p /app/reports

# Set executable permissions
RUN chmod +x /app/inspection-tool

ENTRYPOINT ["/app/inspection-tool"]
CMD ["--help"]
