# Stage 1: Build the Go binary
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app .

# Stage 2: Create a lightweight image to run the Go binary
FROM alpine:3.18

# Set up a working directory
WORKDIR /app

# Copy the built Go binary from the builder image
COPY --from=builder /go/bin/app /app/app

# Expose the port where the Go app runs
EXPOSE 8080

# Run the Go binary
CMD ["/app/app"]




