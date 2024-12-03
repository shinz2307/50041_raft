# Use the official Golang image as the base
FROM golang:1.23.2

# Set the working directory inside the container
WORKDIR /app

# Copy the application code
COPY . .

# Build the Go application
RUN go run main.go -single-rpc

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./app"]
