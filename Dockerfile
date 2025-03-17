# Build stage
FROM golang:1.23-alpine AS builder 

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM golang:1.23-alpine

# Set working directory
WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata bash

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Create uploads directory with subdirectories
RUN mkdir -p /app/uploads/products

# Set permissions for uploads directory
RUN chmod -R 755 /app/uploads

# Run the application
CMD ["./main"] 