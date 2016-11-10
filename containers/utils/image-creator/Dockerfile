FROM ubuntu:16.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y && \
    apt-get install -y parted kpartx curl qemu-utils dosfstools m4 pkg-config && \
    apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY image-creator /

ENTRYPOINT ["/image-creator"]