Scyd is a REST API + web ui built with Go and Vue.js to wrap multiple music downloaders.

It currently wraps yt-dlp and streamrip to support the current platforms:

- YouTube
- SoundCloud
- Qobuz
- Deezer

## Features

- REST API: call the api directly to download songs and playlist, create automated workflows and more
- web ui: paste links in the ui and follow the download progress in real-time thanks to websockets
- config: configure how the downloads behave with a simple yaml config file
- organizer: when downloads finish, automatically organize songs in the right output directory so servers like Jellyfin or Plex can pick them up
- pwa: the web UI is a PWA (it has to be served over https), this means you can install it on your device and enable the share functionality. Example: when browsing YouTube, you might want to download a song, you can then share the song you like with the share button, select the app you want to share with -> in this case scyd -> the app picks up the song's link and immediately starts a download!

## Quick start with docker

**Docker run**

```bash
docker run image_name:latest \
    # WEB UI + REST API port, should be served by a reverse proxy on https
    -p 3000:3000 \
    # directory when songs are downloaded initially
    -v ./downloads:/downloads \
    # output dir where songs are placed on download success, this could be your Jellyfin library dir
    -v ./output:/output \
    # please create a config.yaml file in this directory
    -v ./config:/app/config
```

**Docker compose**

```yaml
services:
  scyd:
    image: image_name:latest
    container_name: scyd
    ports:
      - "3000:3000" # WEB UI + REST API port, should be served by a reverse proxy on https
    volumes:
      - ./downloads:/downloads # directory when songs are downloaded initially
      - ./output:/output # output dir where songs are placed on download success, this could be your Jellyfin library dir
      - ./config:/app/config # please create a config.yaml file in this directory
    restart: unless-stopped
```
