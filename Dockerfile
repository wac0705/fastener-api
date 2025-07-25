# --- Stage 1: The Builder ---
# This stage builds the Go binary.
FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

# Install git, which is needed for go modules
RUN apk add --no-cache git

# Copy the module files first
COPY go.mod go.sum ./

# Download all dependencies. This command leverages Docker's layer caching.
# It will only re-run if go.mod or go.sum change.
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application, creating a statically-linked binary
# CGO_ENABLED=0 is important for creating a binary that can run on a minimal base image like scratch or alpine
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /main .


# --- Stage 2: The Final Image ---
# This stage creates the tiny final image for production.
FROM alpine:latest

# It's good practice to run as a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy only the compiled binary from the builder stage
COPY --from=builder /main /main

# Expose the port the app runs on
EXPOSE 3001

# Set the entrypoint for the container
ENTRYPOINT ["/main"]
