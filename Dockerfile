# Use official Golang image
FROM golang:1.23.5

# Set working directory inside container
WORKDIR /app

# Copy everything into the container
COPY . .

# Download dependencies
RUN go mod tidy

# Expose the application port
EXPOSE 3000

# Run the Go application
CMD ["go", "run", "main.go"]