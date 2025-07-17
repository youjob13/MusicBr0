FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (no CGO needed for modernc.org/sqlite)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o music-bot .

# Start a new stage from scratch
FROM alpine:latest

# Install ca-certificates and tzdata (sqlite not needed with pure Go driver)
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/music-bot .

# Create data directory for database
RUN mkdir -p /root/data

# Set timezone
ENV TZ=UTC

# Expose port (Heroku assigns PORT dynamically) 
EXPOSE 8080

# Run the binary
CMD ["./music-bot"] 