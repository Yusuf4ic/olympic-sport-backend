# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Build the main API server and the seeder CLI
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/seed ./cmd/seed

# Stage 2: Minimal runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the compiled binaries
COPY --from=builder /app/api /app/api
COPY --from=builder /app/seed /app/seed

# Copy migrations only (no .env — Render injects env vars directly)
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/api"]
