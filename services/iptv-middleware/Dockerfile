# services/iptv-middleware/Dockerfile
# FFmpeg + Go transcoder with NVIDIA GPU support
FROM nvidia/cuda:12.2-devel-ubuntu22.04 AS builder

RUN apt-get update && apt-get install -y     golang-1.22     git     pkg-config     && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/lib/go-1.22/bin:${PATH}"

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags nvml     -o /bin/transcoder     ./transcode-engine/cmd/transcoder

# Runtime stage with FFmpeg
FROM jrottenberg/ffmpeg:6.0-nvidia

RUN apt-get update && apt-get install -y     ca-certificates     && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/transcoder /usr/local/bin/transcoder

# Create stream output directory
RUN mkdir -p /var/streams

EXPOSE 8081 9090
ENTRYPOINT ["/usr/local/bin/transcoder"]
