FROM golang:1.22-alpine

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application's source code
COPY . .

# Build the application
RUN go build -o krantorbox

# Run the application
CMD ["/app/krantorbox"]