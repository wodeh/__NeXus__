# Multi-stage build for Session Manager
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build     -ldflags="-w -s -X main.ServiceVersion=$(git describe --tags --always)"     -o /build/session-manager     ./main.go

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/session-manager /session-manager

EXPOSE 8081

USER 65534:65534

ENTRYPOINT ["/session-manager"]
