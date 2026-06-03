# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/dashboard main.go

# Stage 2: Build a small image
FROM alpine:latest

WORKDIR /app

# Add ca-certificates in case we need them for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/dashboard /app/dashboard

# Copy static assets and templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/templates /app/templates

# Create an empty config.yaml to be mounted over later, or provide the default template
COPY --from=builder /app/config.yaml /app/config.yaml

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/dashboard"]
