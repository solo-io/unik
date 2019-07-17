FROM ubuntu:16.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y && \
  apt-get install -y parted kpartx curl qemu-utils dosfstools opam m4 pkg-config wget &&\
  apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*  &&\
  wget -O bubblewrap.deb http://lug.mtu.edu/ubuntu/pool/main/b/bubblewrap/bubblewrap_0.3.1-2_amd64.deb &&\
  dpkg -i bubblewrap.deb &&\
  #opam init --disable-sandboxing --reinit -y && opam switch 4.06.0
  wget -O /usr/local/bin/opam https://github.com/ocaml/opam/releases/download/2.0.1/opam-2.0.1-x86_64-linux &&\
  chmod a+x /usr/local/bin/opam &&\
  yes '' | opam init --disable-sandboxing --reinit -y && yes '' | opam switch create 4.06.0 &&\
  cd /tmp/ && \
  yes '' | opam source fat-filesystem --dir ocaml-fat && \
  cd /tmp/ocaml-fat && \
  yes '' | opam pin add fat-filesystem . -n -y && \
  yes '' | opam install fat-filesystem --verbose -y

ENV CAML_LD_LIBRARY_PATH="/root/.opam/system/lib/stublibs:/usr/lib/ocaml/stublibs"
ENV MANPATH="/root/.opam/system/man:"
ENV PERL5LIB="/root/.opam/system/lib/perl5"
ENV OCAML_TOPLEVEL_PATH="/root/.opam/system/lib/toplevel"
ENV PATH="/root/.opam/system/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

COPY image-creator /

ENTRYPOINT ["/image-creator"]
