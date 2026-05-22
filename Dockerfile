FROM --platform=linux/arm64 golang:1.26-bookworm

WORKDIR /app

# Add Raspberry Pi apt archive — provides gstreamer1.0-libcamera and libcamera-ipa
RUN apt-get update && apt-get install -y --no-install-recommends gnupg2 curl \
    && curl -fsSL https://archive.raspberrypi.com/debian/raspberrypi.gpg.key \
       | gpg --dearmor -o /etc/apt/trusted.gpg.d/raspberrypi.gpg \
    && echo "deb http://archive.raspberrypi.com/debian/ bookworm main" \
       > /etc/apt/sources.list.d/raspi.list \
    && printf 'Package: *\nPin: origin "archive.raspberrypi.com"\nPin-Priority: 1001\n' \
       > /etc/apt/preferences.d/raspberrypi \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y --no-install-recommends \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-libcamera \
    libcamera-ipa \
    rpicam-apps \
    && rm -rf /var/lib/apt/lists/*

ADD . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /rpi-security-cam .

EXPOSE 8080
CMD ["/rpi-security-cam"]
