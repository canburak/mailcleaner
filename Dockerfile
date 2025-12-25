# Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build backend
FROM golang:1.21-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o mailcleaner-server ./cmd/server
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o mailcleaner ./cmd/mailcleaner

# Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# Create non-root user
RUN addgroup -g 1000 mailcleaner && \
    adduser -u 1000 -G mailcleaner -s /bin/sh -D mailcleaner

WORKDIR /app

# Copy binaries
COPY --from=backend-builder /app/mailcleaner-server /app/
COPY --from=backend-builder /app/mailcleaner /app/

# Copy frontend
COPY --from=frontend-builder /app/web/dist /app/web/dist

# Create data directory
RUN mkdir -p /data && chown -R mailcleaner:mailcleaner /data /app

USER mailcleaner

# Environment variables
ENV MAILCLEANER_DB=/data/data.db
ENV MAILCLEANER_PORT=8080

EXPOSE 8080

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/accounts || exit 1

ENTRYPOINT ["/app/mailcleaner-server"]
CMD ["-port", "8080", "-db", "/data/data.db", "-static", "/app/web/dist"]
