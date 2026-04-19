# Production stage
FROM golang:alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /src

# Copy go-shared first (it's a local dependency)
COPY backend/go-shared /src/backend/go-shared

# Copy gateway
COPY gateway /src/gateway

# Build gateway
WORKDIR /src/gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gateway cmd/server/main.go

# Final stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/gateway .

# Expose port
EXPOSE 8000

# Start gateway
CMD ["./gateway"]
