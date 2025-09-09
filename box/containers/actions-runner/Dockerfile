FROM summerwind/actions-runner:v2.328.0-ubuntu-22.04

ARG BUILD_DATE

LABEL org.opencontainers.image.base.name="docker.io/summerwind/actions-runner:v2.328.0-ubuntu-22.04" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/younsl/box/tree/main/box/containers/actions-runner" \ 
      org.opencontainers.image.title="actions-runner" \
      org.opencontainers.image.version="0.2.0"

USER root

# Add additional APT sources for better availability
COPY additional-sources.list /tmp/
RUN cat /tmp/additional-sources.list >> /etc/apt/sources.list && \
    rm /tmp/additional-sources.list

# Install make and other build essentials
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    make \
    && rm -rf /var/lib/apt/lists/*

USER runner
