# Use the official Go image
FROM golang:1.23.2

# Set the working directory in the container
WORKDIR /app

# Copy the Go script into the container
COPY . /app

# Expose the application port (optional if your Go script binds to a specific port)
EXPOSE 8080

# Command to run your Go script
CMD ["go", "run", "main.go", "-single-rpc"]
