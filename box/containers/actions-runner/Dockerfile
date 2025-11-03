FROM summerwind/actions-runner:v2.329.0-ubuntu-24.04

ARG BUILD_DATE

LABEL org.opencontainers.image.base.name="docker.io/summerwind/actions-runner:v2.329.0-ubuntu-24.04" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/younsl/box/tree/main/box/containers/actions-runner" \
      org.opencontainers.image.title="actions-runner" \
      org.opencontainers.image.version="0.4.0"

USER root

# Add additional APT sources for better availability
COPY kakao-mirror.sources /etc/apt/sources.list.d/kakao-mirror.sources

# Install make and other build essentials
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    make \
    && rm -rf /var/lib/apt/lists/*

USER runner
