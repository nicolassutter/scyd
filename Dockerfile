FROM python:3-trixie AS runtime
WORKDIR /app
RUN pip3 install git+https://github.com/nathom/streamrip.git@dev
RUN apt update && apt install -y ffmpeg wget
# install yt-dlp
RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

FROM golang:1.25-alpine AS build-go
WORKDIR /app
COPY ./backend /app
RUN go build -o bin/scyd main.go

FROM oven/bun AS build-frontend
COPY ./frontend /app
WORKDIR /app
RUN bun install
RUN bun run generate

FROM runtime
COPY --from=build-go /app/bin/scyd /app/scyd
COPY --from=build-frontend /app/.output/public /app/public
CMD ["bash", "-c", "cd /app && GO_ENV=production ./scyd"]
