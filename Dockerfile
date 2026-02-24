# syntax=docker/dockerfile:1

# Dockerfile для Backend на Go
# Multi-stage build: builder -> runner

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install and run gqlgen for code generation (версия из go.mod)
RUN go install github.com/99designs/gqlgen@v0.17.86 && \
    cd graph && gqlgen generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/graph

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

ENV PORT=4000
ENV GIN_MODE=release
ENV MONGODB_URI=mongodb://mongo:27017/cnpf_feeder
ENV AUTH_SECRET=change_this_to_a_long_random_string

EXPOSE 4000

CMD ["./main"]
