FROM alpine

LABEL org.opencontainers.image.source = "https://github.com/zebbra/snow-exporter"
LABEL org.opencontainers.image.license = "MIT"

COPY snow-exporter /snow-exporter
ENTRYPOINT ["/snow-exporter"]
