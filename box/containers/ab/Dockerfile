# Container image for apache benchmark(ab)
# Usage:
# docker run --rm -it \
#   <IMAGE_NAME:TAG> \
#   -c 10 \
#   -n 100 \
#   http://localhost:8080/
#

FROM alpine:3.22.2
LABEL maintainer="Younsung Lee <cysl@kakao.com>" \
      version="1.0.0" \
      description="Lightweight container image for running apache benchmark(ab)" \
      repository="https://github.com/younsl/box" \
      homepage="https://younsl.github.io" \
      runtime="docker-desktop, podman" \
      build-date="2025-02-01"

RUN apk update && \
    apk add --no-cache \
        apache2-ssl \
        apache2-utils \
        ca-certificates \
        htop

WORKDIR /app

ENTRYPOINT ["ab"]