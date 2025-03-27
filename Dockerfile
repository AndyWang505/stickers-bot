FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o stickers-bot

# Create resources directory
RUN mkdir -p resources

# Run the application
CMD ["./stickers-bot"] 