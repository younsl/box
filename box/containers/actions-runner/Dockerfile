FROM summerwind/actions-runner:v2.328.0-ubuntu-22.04

USER root

# Add additional APT sources for better availability
COPY additional-sources.list /tmp/
RUN cat /tmp/additional-sources.list >> /etc/apt/sources.list && \
    rm /tmp/additional-sources.list && \
    apt-get update

USER runner
