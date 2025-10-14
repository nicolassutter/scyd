# Scyd 🎵

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.0+-4FC08D?logo=vue.js)](https://vuejs.org)

A self-hosted music downloader with REST API and web interface built with Go and Vue.js. Scyd wraps multiple music downloaders to support various platforms and helps you build your personal music library.

## 🎯 Overview

Scyd provides a unified interface for downloading music from multiple platforms using popular tools like yt-dlp and streamrip. Perfect for self-hosters and music enthusiasts who want to create and organize their personal music collections.

### Supported Platforms

- 🎥 **YouTube** - Videos and playlists
- 🎧 **SoundCloud** - Tracks and sets
- 🎼 **Qobuz** - High-quality audio
- 🎵 **Deezer** - Music streaming platform

## ✨ Features

- **🔓 Open Source & Self-Hosted** - Complete control over your music downloads
- **🔌 REST API** - Integrate with automation tools and custom workflows
- **🖥️ Web Interface** - User-friendly UI with real-time download progress via WebSockets
- **⚙️ Configurable** - Customize download behavior with YAML configuration
- **📁 Auto-Organization** - Automatically organize downloaded music for media servers (Jellyfin, Plex)
- **🪝 Hooks** - Execute custom commands on download completion or errors
- **📱 Progressive Web App (PWA)** - Install on devices and use share functionality from other apps
- **🔗 Share Integration** - Download music directly from your devices's share menu

### PWA Share Feature

When served over HTTPS, Scyd can be installed as a PWA on your devices. This enables powerful share functionality - simply share a YouTube link or SoundCloud track from your device, select Scyd, and the download starts automatically!

## 🚀 Quick Start

### Prerequisites

- Docker and/or Docker Compose
- HTTPS setup (required for PWA features)
- Basic understanding of YAML configuration

### Docker Run

```bash
docker run -d \
  --name scyd \
  -p 3000:3000 \
  -v ./downloads:/downloads \
  -v ./output:/output \
  -v ./config:/app/config \
  --restart unless-stopped \
  ghcr.io/nicolassutter/scyd:latest
```

### Docker Compose (Recommended)

Create a `compose.yaml` file:

```yaml
services:
  scyd:
    image: ghcr.io/nicolassutter/scyd:latest
    container_name: scyd
    ports:
      - "3000:3000" # Web UI + REST API
    volumes:
      - ./downloads:/downloads # Temporary download directory
      - ./output:/output # Final organized music library
      - ./config:/app/config # Configuration directory
    restart: unless-stopped
    environment:
      - TZ=UTC # Set your timezone
```

Run with:

```bash
docker-compose up -d
```

## 📁 Directory Structure

| Path         | Purpose                                                      |
| ------------ | ------------------------------------------------------------ |
| `/downloads` | Temporary storage for downloads in progress                  |
| `/output`    | Final organized music library (point your media server here) |
| `/config`    | Configuration files directory                                |

## ⚙️ Configuration

Create a `config.yaml` file in your config directory.

Example minimal configuration:

```yaml
users:
  username1:
    password_hash: "<bcrpt hashed password>"
  username2:
    password_hash: "<bcrpt hashed password>"

sort_after_download: true # can disable automatic sorting

hooks:
  on_error: curl https://your-webhook-url/error
  on_download_complete: curl https://your-webhook-url/success
```

## 🔧 Development

### Local Development Setup

Prequisites: bun, Go, Docker, air, make

1. **Clone the repository**

   ```bash
   git clone https://github.com/nicolassutter/scyd.git
   cd scyd
   ```

2. **Backend (Go)**

   ```bash
   cd backend
   go mod download
   ```

3. **Frontend (Vue.js/Nuxt)**

   ```bash
   cd frontend
   bun install
   ```

4. **Run Backend and Frontend**

   ```bash
   make dev
   ```

## 📖 API Documentation

The REST API provides automatic documentation at `/docs`. Endpoints are protected with cookie-based authentication.
You would first call the `/login` endpoint to obtain a session cookie and then send this cookie with subsequent requests.

## 🤝 Contributing

Do not hesitate to open issues or submit pull requests. Contributions are welcome!

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- [streamrip](https://github.com/nathom/streamrip)
- [Go Fiber](https://gofiber.io/)
- [Nuxt.js](https://nuxt.com/)

## ⚠️ Disclaimer

This tool is for educational and personal use only. Please respect copyright laws and the terms of service of the platforms you download from. Users are responsible for ensuring their usage complies with applicable laws and platform policies.
