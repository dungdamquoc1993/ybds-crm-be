# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the .env file
COPY .env .

# Create uploads directory
RUN mkdir -p uploads

# Expose the port
EXPOSE 3000

# Run the application
CMD ["./main"] 