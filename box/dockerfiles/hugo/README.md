# hugo

Hugo server for local development

## Usage

```bash
docker build -t hugo:dev .

SOURCE=/path/to/your/hugo/blog/repo # or `SOURCE=$(pwd)` if you are in the repo directory
docker run -d --name hugo -p 1313:1313 -v ${SOURCE}:/app hugo:dev
```

You can check the container status using `docker ps`.

```bash
docker ps
```

After the container starts successfully, open your browser and go to [http://localhost:1313](http://localhost:1313) to view your hugo blog in real-time and start local development.

## Important Notes

When using Alpine Linux as base image, you must install `libc6-compat` and `g++` packages using apk(alpine package manager) to resolve the `/bin/sh: /usr/local/bin/hugo: not found` error during container build.

Add the following commands in your dockerfile if you are using `alpine` image:

```dockerfile
FROM alpine:3.20.0

RUN apk update && \
    apk add --no-cache \
        git \
        wget \
        libc6-compat \
        g++ && \
    # ... omitted for brevity ...
```
