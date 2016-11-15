FROM ubuntu:16.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y && \
    apt-get install -y parted kpartx curl qemu-utils dosfstools opam m4 pkg-config && \
    apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN opam init -y && cd /tmp/ && \
    opam remote add mirage-dev https://github.com/mirage/mirage-dev.git && \
    opam update && \
    git clone https://github.com/mirage/ocaml-fat && \
    cd /tmp/ocaml-fat && \
    opam pin add ocaml-fat . -n -y && \
    opam install ocaml-fat --verbose

ENV CAML_LD_LIBRARY_PATH="/root/.opam/system/lib/stublibs:/usr/lib/ocaml/stublibs"
ENV MANPATH="/root/.opam/system/man:"
ENV PERL5LIB="/root/.opam/system/lib/perl5"
ENV OCAML_TOPLEVEL_PATH="/root/.opam/system/lib/toplevel"
ENV PATH="/root/.opam/system/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

COPY image-creator /

ENTRYPOINT ["/image-creator"]