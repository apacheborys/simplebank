# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy only the necessary files for the Golang application
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
