# Stage 1: Build Go application
FROM golang:1.23 AS go-builder

WORKDIR /app

# Copy Go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o tryyourgyan main.go

# Stage 2: Set up Python and final image
FROM python:3.10-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy Go binary from go-builder stage
COPY --from=go-builder /app/tryyourgyan .

# Copy quizlogic directory
COPY quizlogic /app/quizlogic

# Set up Python virtual environment
RUN python -m venv /app/quizlogic/venv
RUN /app/quizlogic/venv/bin/pip install --no-cache-dir -r /app/quizlogic/requirements.txt

# Ensure app.py is executable
RUN chmod +x /app/quizlogic/app.py

# Expose port
EXPOSE 8080

# Command to run Go application
CMD ["./tryyourgyan"]