# --------------------
# Stage 1: Build the app
# --------------------
    FROM golang:1.23.0 AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # ✅ Build a static binary
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /app/server ./main.go
    
    
    # --------------------
    # Stage 2: Run the app
    # --------------------
    FROM alpine:latest
    
    RUN apk --no-cache add ca-certificates
    
    WORKDIR /root/
    
    # ✅ Copy the static binary
    COPY --from=builder /app/server ./server
    RUN chmod +x ./server
    
    EXPOSE 8080
    
    CMD ["./server"]
    