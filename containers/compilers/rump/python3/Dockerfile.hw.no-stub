FROM projectunik/compilers-rump-go-hw:a92f4aa53a414bbf

RUN apt-get update
RUN apt-get install -y libsqlite3-dev libssl-dev
RUN mkdir -p /opt/python3
RUN cd /opt/python3 && git clone https://github.com/rumpkernel/rumprun-packages
RUN cd /opt/python3/rumprun-packages/python3 && \
    cp ../config.mk.dist ../config.mk && \
    perl -pi -e 's/RUMPRUN_TOOLCHAIN_TUPLE=/RUMPRUN_TOOLCHAIN_TUPLE=x86_64-rumprun-netbsd/g' ../config.mk && \
    sed -i '/\$(RUMPRUN_GENISOIMAGE) -o images\/python.iso build\/pythondist\/lib\/python3.5/d' Makefile && \
    export DESTDIR= && \
    make
RUN mkdir -p /python/lib
RUN cp -r /opt/python3/rumprun-packages/python3/build/pythondist/lib/python3.5 /python/lib/

WORKDIR /opt

ENV RUMP_BAKE=hw_generic

RUN rumprun-bake $RUMP_BAKE /program.bin /opt/python3/rumprun-packages/python3/build/python

COPY python-wrapper /build/python-wrapper/

# RUN LIKE THIS: docker run --rm -v /path/to/code:/opt/code -e MAIN_FILE=main_file.js -e BOOTSTRAP_TYPE=ec2|udp projectunik/compilers-rump-python-hw
CMD set -x && \
    (if [ -z "$MAIN_FILE" ]; then echo "Need to set MAIN_FILE"; exit 1; fi) && \
    cp /build/python-wrapper/python-wrapper-no-stub.py /opt/code/python-wrapper.py && \
    mkdir -p /opt/code/python/lib && \
    perl -pi -e 's/import main.py/import $ENV{MAIN_FILE}/g' /opt/code/python-wrapper.py && \
    cp -r /python/lib/* /opt/code/python/lib/ && \
    cp /program.bin /opt/code/program.bin
