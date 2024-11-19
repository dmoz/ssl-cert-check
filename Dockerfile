# Use the official Go image as a base
FROM golang:alpine

# Set the working directory to /app
WORKDIR /app

# Copy the Go mod files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the application code
COPY . .

# Build the application
RUN go build -o tls-monitor main.go

# Expose the port (not needed in this case, but good practice)
# EXPOSE 443

# Run the command to start the application
CMD ["./tls-monitor"]