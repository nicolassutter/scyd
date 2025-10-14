# Multi-stage build optimized for GitHub Actions
FROM python:3.14.0 AS runtime-base
WORKDIR /app
# Install system dependencies first (they change less frequently)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install Python packages
RUN pip3 install --no-cache-dir streamrip --upgrade

# Install yt-dlp
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

FROM golang:1.25-bookworm AS build-go
WORKDIR /app
COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download
COPY ./backend .
RUN CGO_ENABLED=1 go build -o bin/scyd main.go

FROM oven/bun AS build-frontend
WORKDIR /app
COPY ./frontend/package.json ./frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY ./frontend .
RUN bun run generate

FROM runtime-base
COPY --from=build-go /app/bin/scyd /app/scyd
COPY --from=build-frontend /app/.output/public /app/public
CMD ["bash", "-c", "cd /app && GO_ENV=production ./scyd"]