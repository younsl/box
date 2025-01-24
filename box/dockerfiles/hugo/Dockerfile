FROM alpine:3.20.0

LABEL maintainer="Younsung Lee <cysl@kakao.com>" \
      version="1.0" \
      description="Docker image for running Hugo blog server" \
      runtime="docker-desktop" \
      build-date="2025-01-25"

ARG HUGO_VERSION=0.142.0
ARG ARCH=arm64

# IMPORTANT:
# To fix the issue "/bin/sh: /usr/local/bin/hugo: not found" during the container build process, 
# you must install the libc6-compat and g++ packages in the alpine image.
# See: https://github.com/gohugoio/hugo/issues/4939#issuecomment-484956006
RUN apk update && \
    apk add --no-cache \
        git \
        wget \
        libc6-compat \
        g++ && \
    wget -O hugo_extended_${HUGO_VERSION}-linux-${ARCH}.tar.gz https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-${ARCH}.tar.gz && \
    tar -xzvf hugo_extended_${HUGO_VERSION}-linux-${ARCH}.tar.gz && \
    rm hugo_extended_${HUGO_VERSION}-linux-${ARCH}.tar.gz && \
    mv hugo /usr/local/bin/ && \
    chmod +x /usr/local/bin/hugo && \
    ls -l /usr/local/bin/ && \
    /usr/local/bin/hugo version

WORKDIR /app
VOLUME /app

EXPOSE 1313

ENTRYPOINT ["hugo", "server"]
CMD ["-t", "nostyleplease", "--bind", "0.0.0.0"]
