FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

EXPOSE 8080