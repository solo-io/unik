FROM ubuntu:16.04

RUN apt-get update -y  &&  apt-get install libxen-dev curl git build-essential -y &&  apt-get clean -y &&  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV RUMPDIR=/usr/local

RUN cd /opt && \
    git clone https://github.com/rumpkernel/rumprun
RUN cd /opt/rumprun && git checkout 16a7c6eb44523c60ea714a0ec2c7ea6ab3c8fb02
RUN cd /opt/rumprun && git submodule update --init


VOLUME /opt/code
WORKDIR /opt/code
