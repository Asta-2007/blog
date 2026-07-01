# =========================
# Stage 1: Build
# =========================
FROM golang:1.25.10-alpine AS builder

WORKDIR /app

# Install git (required for some Go modules)
RUN apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o server .

# =========================
# Stage 2: Runtime
# =========================
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server .

# Railway akan menyediakan PORT saat runtime
ENV PORT=8080

EXPOSE 8080

CMD ["./server"]