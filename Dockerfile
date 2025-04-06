# --------------------
# Stage 1: Build the app
# --------------------
    FROM golang:1.23.0 AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    RUN GOOS=linux GOARCH=amd64 go build -v -o /app/server ./main.go
    
    
    # --------------------
    # Stage 2: Run the app
    # --------------------
    FROM debian:bullseye-slim
    
    RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
    
    WORKDIR /root/
    
    COPY --from=builder /app/server ./server
    RUN chmod +x ./server
    
    EXPOSE 8080
    
    CMD ["./server"]
    