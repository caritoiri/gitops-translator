# =========================================================
# Stage 1: Build the static Go binary
# =========================================================
FROM golang:1.26.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy dependency definitions and download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Compile high-performance, statically linked Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gitops-translator main.go

# =========================================================
# Stage 2: Serve the binary from an ultra-lightweight image
# =========================================================
FROM alpine:3.19

# Set secure system environment
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled static binary from builder
COPY --from=builder /app/gitops-translator /app/gitops-translator

# Default listening port
EXPOSE 8080

# Run the statically linked binary natively
ENTRYPOINT ["/app/gitops-translator"]
