# services/backend-core/Dockerfile
# Multi-stage build for Go backend
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build     -ldflags='-w -s -extldflags "-static"'     -a -installsuffix cgo     -o /bin/backend-core     ./cmd/server

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bin/backend-core /backend-core

EXPOSE 8080 8081 9090
ENTRYPOINT ["/backend-core"]
