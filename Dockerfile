# Build stage
FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin/webserver web/main.go

# Final stage
FROM alpine:3.23
RUN apk add --no-cache ca-certificates=20251003-r0 curl=8.17.0-r1
WORKDIR /app
COPY --from=builder /app/bin/webserver .
EXPOSE 8080
ENV LANG=C.UTF-8
ENV GIN_MODE=release
CMD ["./webserver"]
