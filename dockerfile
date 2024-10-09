FROM golang:1.22.5

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy application source code and .env file
COPY cmd/ ./cmd/
COPY db/ ./db/
COPY templates/ ./templates/
COPY .env ./.env

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Set the command to run the application
CMD ["./main"]