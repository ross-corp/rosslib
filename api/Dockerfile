FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 go build -o /pocketbase main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /pocketbase /app/pocketbase

EXPOSE 8090

# Default command to start PocketBase
CMD ["/app/pocketbase", "serve", "--http=0.0.0.0:8090"]
