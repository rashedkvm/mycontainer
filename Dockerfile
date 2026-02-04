# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server main.go

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY static/ ./static/
EXPOSE 8080
CMD ["./server"]
