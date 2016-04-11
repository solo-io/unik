FROM avsm/docker-opam:ubuntu-trusty-4.01.0

RUN sudo -u opam sh -c "cd /home/opam/ && opam update -u -y && opam installext -y mirage"

VOLUME  /opt/code
WORKDIR /opt/code

CMD mirage configure --xen && make
