# Stage 1: Build the Go binary
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application statically compiled
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/dhiarhome main.go

# Stage 2: Build a small image
FROM alpine:latest

WORKDIR /app

# Add ca-certificates in case we need them for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/dhiarhome /app/dhiarhome

# Copy static assets and templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/templates /app/templates

# Create data directory for favicon cache and todo persistence
RUN mkdir -p /app/data/icons

# Create an empty config.yaml from the example template.
# IMPORTANT: Mount your own config.yaml at runtime via volume:
#   docker run -v ./config.yaml:/app/config.yaml dhiarhome
COPY --from=builder /app/config-example.yaml /app/config.yaml

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/dhiarhome"]
