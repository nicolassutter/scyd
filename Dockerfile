FROM alpine:latest AS runtime-base
WORKDIR /app
# Install system dependencies first (they change less frequently)
RUN apk add --no-cache \
    ffmpeg \
    wget

# Install yt-dlp for musl libc (Alpine) | would use the glibc version on Debian/Ubuntu
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_musllinux -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

FROM golang:1.25-bookworm AS build-go
WORKDIR /app
COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download
COPY ./backend .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o bin/scyd main.go

FROM oven/bun AS build-frontend
WORKDIR /app
COPY ./frontend/package.json ./frontend/bun.lock ./
RUN bun install --frozen-lockfile
COPY ./frontend .
RUN bun run generate

FROM runtime-base
COPY --from=build-go /app/bin/scyd /app/scyd
COPY --from=build-frontend /app/.output/public /app/public
ENV GO_ENV=production
CMD ["/bin/sh", "-c", "./scyd"]