FROM ubuntu:22.04

LABEL maintainer "Younsung Lee <cysl@kakao.com>"
LABEL version="1.0"
LABEL description "Terraform console machine"

RUN apt-get update && \
    apt-get install -y \
            curl \
            git \
            unzip \
            vim \
            tree \
    && curl -L https://raw.githubusercontent.com/warrensbox/terraform-switcher/release/install.sh | bash \
    && curl -L https://raw.githubusercontent.com/warrensbox/tgswitch/release/install.sh | bash \
    && curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" \
    && unzip awscliv2.zip \
    && ./aws/install

ENV TF_PLUGIN_CACHE_DIR=/root/.terraform.d/plugin-cache
