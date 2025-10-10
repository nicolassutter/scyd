FROM python:3-trixie

RUN pip3 install git+https://github.com/nathom/streamrip.git@dev

RUN apt update && apt install -y ffmpeg wget

# install yt-dlp

RUN wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp
