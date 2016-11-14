FROM ubuntu:16.04
MAINTAINER includeos.org
RUN \
  apt-get update && \
  apt-get install -y bc git lsb-release sudo
#clone & checkout dev
RUN cd ~ && pwd && \
  git clone https://github.com/hioa-cs/IncludeOS.git && \
  cd IncludeOS && \
  git checkout 0681f147661f11b51bd3783349fece93d958fdea && \
  git fetch --tags

#Dependencies for unik.cpp
RUN mkdir /root/IncludeOS/src/lib/ && \
    cd /root/IncludeOS/src/lib/ && \
    git clone https://github.com/includeos/mana && \
    git clone https://github.com/includeos/json && \
    cd json && \
    git submodule update --init

#Patches
COPY patches /tmp/patches
RUN cp -r /tmp/patches/* /root/IncludeOS/

#Install
RUN cd ~ && pwd && \
  cd IncludeOS && \
  /bin/bash ./install.sh
