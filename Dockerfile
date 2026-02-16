# syntax=docker/dockerfile:1

# Dockerfile для Backend на Go
# Multi-stage build: builder -> runner

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Install gqlgen and all its dependencies explicitly
RUN go get -d github.com/99designs/gqlgen@latest && \
    go get golang.org/x/tools/go/packages && \
    go get golang.org/x/tools/go/ast/astutil && \
    go get golang.org/x/tools/imports && \
    go get github.com/goccy/go-yaml && \
    go get github.com/urfave/cli/v3 && \
    go mod tidy

# Copy source code
COPY . .

# Generate GraphQL code
RUN cd graph && \
    go run github.com/99designs/gqlgen@latest generate

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
