# services/ai-platform/Dockerfile
# AI Inference Gateway + Python ML pipelines
FROM python:3.11-slim AS python-builder

RUN apt-get update && apt-get install -y     build-essential     git     && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY ml-pipelines/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM golang:1.22-alpine AS go-builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build     -o /bin/inference-gateway     ./inference-gateway/cmd/gateway

# Final stage
FROM nvidia/cuda:12.2-runtime-ubuntu22.04

RUN apt-get update && apt-get install -y     python3.11     python3-pip     ca-certificates     && rm -rf /var/lib/apt/lists/*

COPY --from=python-builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=go-builder /bin/inference-gateway /usr/local/bin/inference-gateway

WORKDIR /app
COPY ml-pipelines/ ./ml-pipelines/

EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/inference-gateway"]
