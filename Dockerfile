FROM golang:1.26-alpine

ENV PATH="/go/bin:${PATH}"

WORKDIR /app

# Install air and swag for development
RUN go install github.com/air-verse/air@v1.65.1 && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Copy go files for dependency download
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run Air to start the server with hot reload
CMD ["air", "-c", ".air.toml"]