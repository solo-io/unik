FROM ubuntu

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y && \
    apt-get install -y --force-yes git openjdk-7-jdk curl && \
    apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN curl https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz | sudo tar xz -C /usr/local && sudo mv /usr/local/go /usr/local/go1.5 && sudo ln -s /usr/local/go1.5 /usr/local/go
ENV GOPATH=$HOME/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$GOBIN:/usr/local/go/bin:$PATH

RUN mkdir -p $GOPATH/src/github.com/vmware
RUN cd $GOPATH/src/github.com/vmware && \
    git clone https://github.com/vmware/govmomi && \
    cd govmomi/govc && \
    go get ./... && \
    go install

COPY target/vsphere-client-1.0-SNAPSHOT-jar-with-dependencies.jar /vsphere-client.jar

#run with "java -jar /vsphere-client.jar"
#or
#"govc [command]"