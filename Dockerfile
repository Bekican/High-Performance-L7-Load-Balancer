# Stage 1: Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
# Note: go-sqlite (glebarez) is pure Go, so CGO_ENABLED=0 is fine.
RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer ./cmd/lb/main.go

# Stage 2: Run Stage (Distroless / Scratch / Alpine)
FROM alpine:latest

WORKDIR /root/

# Add certificates for HTTPS calls if needed (though we use raw TCP mostly)
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/loadbalancer .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/dashboard.html .
# Only copy html if it exists, otherwise mkdir
RUN mkdir -p html
COPY --from=builder /app/html ./html

# Expose ports
EXPOSE 8080 9091 443

# Run
CMD ["./loadbalancer"]
