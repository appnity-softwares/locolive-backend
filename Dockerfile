# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies required for building (if any)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime Stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
# curl is needed for downloading migrate
RUN apk add --no-cache ca-certificates curl

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/bin/migrate

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy configuration and migration files
# Viper looks for app.env in the current directory or config paths
# We will use environment variables in Docker Compose, but copying app.env 
# as a fallback or template is good practice.
# Copy app.env.example as app.env so Viper has a file to read
# Actual values will be overridden by environment variables injected by Render
COPY app.env.example app.env
COPY db/migrations ./db/migrations
COPY start.sh .

# Ensure start script is executable
RUN chmod +x start.sh

EXPOSE 8080

ENTRYPOINT [ "/app/start.sh" ]
CMD [ "/app/server" ]
