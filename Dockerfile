# --------------------
# Stage 1: Build the app
# --------------------
    FROM golang:1.23.0 AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # Build binary
    RUN go build -o /app/server ./main.go
    
    
    # --------------------
    # Stage 2: Run the app
    # --------------------
    FROM alpine:latest
    
    RUN apk --no-cache add ca-certificates
    
    WORKDIR /root/
    
    COPY --from=builder /app/server .
    
    EXPOSE 8080
    
    CMD ["./server"]
    