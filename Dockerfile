# Stage 1: Build the Go application
FROM golang:1.22 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Stage 2: Run the Go application
FROM golang:1.22

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary and config file from the builder stage
COPY --from=builder /app/config.toml .
COPY --from=builder /app/main .

# Expose port 9090 to the outside world
EXPOSE 9090

# Command to run the executable
CMD ["./main"]