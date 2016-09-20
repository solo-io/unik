FROM projectunik/compilers-rump-go-hw:a92f4aa53a414bbf

ENV RUMP_BAKE=hw_generic

COPY stub /build/stub/

RUN set -x && cd /build/stub/ && \
    CC=x86_64-rumprun-netbsd-gcc CGO_ENABLED=1 GOOS=rumprun /usr/local/go/bin/go build -buildmode=c-archive -v -a -x  *.go && \
    RUMPRUN_STUBLINK=succeed x86_64-rumprun-netbsd-gcc -g -o /build/stub/stub mainstub.c $(find . -name "*.a")

VOLUME /opt/code

# RUN LIKE THIS: docker run --rm -v /path/to/code:/opt/code -e BINARY_NAME=program projectunik/compilers-rump-c-hw
CMD set -x && \
    (if [ -z "BINARY_NAME" ]; then echo "Need to set MAIN_FILE"; exit 1; fi) && \
    cd /opt/code && make CC=x86_64-rumprun-netbsd-gcc && \
    rumprun-bake $RUMP_BAKE /opt/code/program.bin /build/stub/stub /opt/code/$BINARY_NAME
