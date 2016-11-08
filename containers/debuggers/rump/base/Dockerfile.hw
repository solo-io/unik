FROM projectunik/compilers-rump-base-hw:8cd85d4a7ee1009b

RUN apt-get update -y
RUN apt-get install -y --force-yes texinfo

RUN curl http://ftp.gnu.org/gnu/gdb/gdb-7.11.tar.gz | tar xz -C /opt/
RUN cd /opt/gdb-7.11/gdb && curl 'https://sourceware.org/bugzilla/attachment.cgi?id=8512&action=diff&collapsed=&headers=1&format=raw' | patch

RUN cd /opt/gdb-7.11 && \
    ./configure --target=x86_64-linux-gnu && \
    make && \
    make install