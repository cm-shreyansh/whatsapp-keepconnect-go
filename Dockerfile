# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Create sessions directory
RUN mkdir -p sessions

# Expose port
EXPOSE 3456

# Run the application
CMD ["./main"]