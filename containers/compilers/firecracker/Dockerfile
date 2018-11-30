FROM golang

RUN cd /tmp && \ 
    wget http://dl-cdn.alpinelinux.org/alpine/v3.8/releases/x86_64/alpine-minirootfs-3.8.1-x86_64.tar.gz

COPY inittab /tmp/overlay/etc/inittab
COPY interfaces /tmp/overlay/etc/network/interfaces
COPY start-script /tmp/overlay/start.sh
COPY resolv.conf  /tmp/overlay/etc/resolv.conf

COPY build-image /build-image
CMD ["/bin/bash", "/build-image"]