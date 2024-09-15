# ab (apache benchmark)
# in a container
#
# Usage:
# docker run --rm -it \
# 	<YOUR_CONTAINER_NAME>
#

FROM alpine:latest
LABEL maintainer "Younsung Lee <cysl@kakao.com>"

RUN apk add --no-cache \
		apache2-ssl \
		apache2-utils \
		ca-certificates \
		htop

ENTRYPOINT [ "ab" ]