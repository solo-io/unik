FROM projectunik/compilers-rump-base-hw:3e0e31ac24ba649f

RUN apt-get update
RUN apt-get install -y pkg-config
RUN apt-get install -y openjdk-8-jdk
RUN apt-get install -y cpio
RUN apt-get install -y mercurial
RUN apt-get install -y unzip
RUN apt-get install -y zip

RUN cd /opt/ && git clone https://github.com/rumpkernel/rumprun-packages
RUN cd /opt/rumprun-packages/openjdk8 && \
    cp ../config.mk.dist ../config.mk && \
    perl -pi -e 's/RUMPRUN_TOOLCHAIN_TUPLE=/RUMPRUN_TOOLCHAIN_TUPLE=x86_64-rumprun-netbsd/g' ../config.mk && \
    perl -pi -e 's/images\/jre.iso images\/jar.ffs//g' Makefile && \
    make

ENV RUMP_BAKE=hw_generic

RUN rumprun-bake $RUMP_BAKE \
    /opt/rumprun-packages/openjdk8/bin/java.bin \
    /opt/rumprun-packages/openjdk8/bin/java

RUN mkdir -p /tmp/build
# Get Jetty for .war builds
WORKDIR /tmp/build
RUN curl -O http://repo1.maven.org/maven2/org/eclipse/jetty/jetty-distribution/9.3.11.v20160721/jetty-distribution-9.3.11.v20160721.tar.gz
RUN tar xvf jetty-distribution-9.3.11.v20160721.tar.gz && rm -rf jetty-distribution-*/demo-base

COPY java-wrapper/target/java-wrapper-1.0-SNAPSHOT-jar-with-dependencies.jar /tmp/build/program.jar

VOLUME /opt/code
COPY create_rump_java_bin.sh /tmp/build/

# RUN LIKE THIS: docker run --rm -v /path/to/code:/opt/code projectunik/compilers-rump-java-hw
CMD bash /tmp/build/create_rump_java_bin.sh