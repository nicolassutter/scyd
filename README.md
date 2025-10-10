## Quick start with docker

```bash
docker run image_name:latest \
    -p 3000:3000 \
    -v ./downloads:/downloads \
    -v ./output:/output \
    -v ./config.yaml:/app/config/config.yaml # completely optional
```
