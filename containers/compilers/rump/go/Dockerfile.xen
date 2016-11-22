FROM projectunik/compilers-rump-base-xen:f841298dae4340f7

RUN curl https://storage.googleapis.com/golang/go1.6.3.linux-amd64.tar.gz | tar xz -C /usr/local && \
    mv /usr/local/go /usr/local/go1.6 && ln -s /usr/local/go1.6 /usr/local/go && \
    cd /tmp && git clone https://github.com/deferpanic/gorump && cd gorump && git checkout f1d676f985f8b337b58ba4b81cf808070798be9b


COPY fixrump.sh /tmp/
COPY patches /tmp/patches/

RUN bash -ex /tmp/fixrump.sh

# fix go needs to be re-visited if we use go > 1.5
COPY fixgo.sh /tmp/
COPY gopatches/ /tmp/gopatches/

RUN cd /tmp/gorump/go1.5/src && \
    bash /tmp/fixgo.sh && \
    CGO_ENABLED=0 GOROOT_BOOTSTRAP=/usr/local/go GOOS=rumprun GOARCH=amd64 ./make.bash && \
    cd /tmp && mv gorump/go1.5/ /usr/local/go-patched && \
    rm /usr/local/go  && \
    ln -s /usr/local/go-patched /usr/local/go

ENV RUMP_BAKE=xen_pv


ENV GOROOT=/usr/local/go
ENV GOPATH=/opt/go
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

COPY stub/ /tmp/build/

# RUN LIKE THIS: docker run --rm -e ROOT_PATH=root_package_path -e BOOTSTRAP_TYPE=ec2|udp|gcloud|nostub -v /path/to/code:/opt/code projectunik/compilers-rump-go-xen
CMD set -x && \
    cp /tmp/build/*.go . && \
    mkdir -p ${GOPATH}/src/${ROOT_PATH} && \
    cp -r ./* ${GOPATH}/src/${ROOT_PATH} && \
    cd ${GOPATH}/src/${ROOT_PATH} && \
    GO15VENDOREXPERIMENT=1 CC=x86_64-rumprun-netbsd-gcc CGO_ENABLED=1 GOOS=rumprun /usr/local/go/bin/go build -tags ${BOOTSTRAP_TYPE} -buildmode=c-archive -v -a -x . && \
    cp /tmp/build/mainstub.c . && \
    RUMPRUN_STUBLINK=succeed x86_64-rumprun-netbsd-gcc -g -o /opt/code/program mainstub.c $(find . -name "*.a") && \
    rm -f /opt/code/mainstub.c /opt/code/gomaincaller.go && \
    bash -x $(which rumprun-bake) $RUMP_BAKE /opt/code/program.bin /opt/code/program