# Multi-stage build for PMS Integration API
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pms-api ./cmd/api

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app

COPY --from=builder /build/pms-api .
COPY --from=builder /build/sql/migrations ./sql/migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["./pms-api"]
