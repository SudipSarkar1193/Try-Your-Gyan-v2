# --------------------
# Stage 1: Build the app
# --------------------
    FROM golang:1.23.0 AS builder

    # Set working directory
    WORKDIR /app
    
    # Copy go mod and sum files
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the source code
    COPY . .
    
    # Build the Go app
    RUN go build -o server main.go
    
    
    # --------------------
    # Stage 2: Run the app
    # --------------------
    FROM alpine:latest
    
    # Install SSL certs (for HTTPS etc.)
    RUN apk --no-cache add ca-certificates
    
    # Set working directory
    WORKDIR /root/
    
    # Copy the built binary
    COPY --from=builder /app/server .
    
    
    # Expose the port the app listens on (default 8080 or whatever in cfg.Addr)
    EXPOSE 8080
    
    # Run the binary
    CMD ["./server"]
    