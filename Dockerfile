# Build stage for backend
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download github.com/gin-contrib/cors github.com/gin-gonic/gin github.com/robfig/cron/v3 gorm.io/driver/postgres gorm.io/gorm github.com/golang-jwt/jwt/v5 golang.org/x/crypto github.com/mmcdole/gofeed || true

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o rss-reader ./backend

# Frontend build stage
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Runtime stage
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/rss-reader .
RUN mkdir -p frontend
COPY --from=frontend-builder /app/dist ./frontend

EXPOSE 8080
CMD ["./rss-reader"]
