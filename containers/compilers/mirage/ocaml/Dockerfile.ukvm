FROM ocaml/opam:ubuntu-16.04_ocaml-4.02.3


RUN opam init -y && opam update -u -y && opam repo add mirage-dev git://github.com/mirage/mirage-dev && \
    opam depext -y -i mirage

# result of "opam config env""
ENV CAML_LD_LIBRARY_PATH="/home/opam/.opam/system/lib/stublibs:/usr/lib/ocaml/stublibs"
ENV MANPATH="/home/opam/.opam/system/man:"
ENV PERL5LIB="/home/opam/.opam/system/lib/perl5"
ENV OCAML_TOPLEVEL_PATH="/home/opam/.opam/system/lib/toplevel"
ENV PATH="/home/opam/.opam/system/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# build a sample app to make sure nothing's broken, and install fat volume tools.
RUN cd /tmp && \
    git clone -b mirage-dev https://github.com/mirage/mirage-skeleton && \
    cd mirage-skeleton/stackv4 && \
    mirage configure -t ukvm && \
    make

VOLUME  /opt/code
WORKDIR /opt/code