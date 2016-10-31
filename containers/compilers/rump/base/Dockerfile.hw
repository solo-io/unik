FROM projectunik/compilers-rump-base-common:2eb72d1b386ce2a4

ENV PLATFORM=hw
ENV BUILDRUMP_EXTRA=

RUN cd /opt/rumprun && \
    ./build-rr.sh -d $RUMPDIR -o ./obj $PLATFORM build -- $BUILDRUMP_EXTRA && \
    ./build-rr.sh -d $RUMPDIR -o ./obj $PLATFORM install

COPY fixrump.sh /tmp/
COPY patches /tmp/patches/

RUN bash -ex /tmp/fixrump.sh

VOLUME /opt/code
WORKDIR /opt/code


# RUN LIKE THIS: docker run --rm -v -ti /path/to/code:/opt/code projectunik/compilers-rump-base-common-hw
