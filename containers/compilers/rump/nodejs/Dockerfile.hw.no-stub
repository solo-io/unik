FROM projectunik/compilers-rump-base-hw:3e0e31ac24ba649f

RUN apt-get update
RUN apt-get install -y python
RUN mkdir -p /opt/nodejs
RUN cd /opt/nodejs && git clone https://github.com/rumpkernel/rumprun-packages
RUN cd /opt/nodejs/rumprun-packages/nodejs && \
    cp ../config.mk.dist ../config.mk && \
    perl -pi -e 's/RUMPRUN_TOOLCHAIN_TUPLE=/RUMPRUN_TOOLCHAIN_TUPLE=x86_64-rumprun-netbsd/g' ../config.mk && \
    make

COPY node-wrapper /opt/node-wrapper/

VOLUME /opt/code
WORKDIR /opt/nodejs/rumprun-packages/nodejs

ENV RUMP_BAKE=hw_generic

RUN rumprun-bake $RUMP_BAKE \
    /opt/nodejs/rumprun-packages/nodejs/build-4.3.0/out/Release/node-default.bin \
    /opt/nodejs/rumprun-packages/nodejs/build-4.3.0/out/Release/node-default

# RUN LIKE THIS: docker run --rm -v /path/to/code:/opt/code projectunik/compilers-rump-nodejs-hw
CMD set -x && \
    (if [ -z "$MAIN_FILE" ]; then echo "Need to set MAIN_FILE"; exit 1; fi) && \
    mv /opt/node-wrapper/node-wrapper-no-stub.js /opt/code/node-wrapper.js && \
    perl -pi -e 's/\/\/CALL_NODE_MAIN_HERE/require("\.\/$ENV{MAIN_FILE}")/g' /opt/code/node-wrapper.js && \
    cp /opt/nodejs/rumprun-packages/nodejs/build-4.3.0/out/Release/node-default.bin /opt/code/program.bin
